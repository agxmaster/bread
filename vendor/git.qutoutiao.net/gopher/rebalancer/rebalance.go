package rebalancer

import (
	"net/url"
	"strconv"
	"sync"
	"time"

	"git.qutoutiao.net/gopher/qudiscovery"
	log "github.com/sirupsen/logrus"

	"git.qutoutiao.net/gopher/rebalancer/degrade"
)

// LoadBalance 负载均衡interface
type LoadBalance interface {
	Upsert(name string, weight int, node interface{}) error
	Next() (interface{}, error)
}

// NewLoadBalanceFunc 新建负载均衡器
type NewLoadBalanceFunc func() LoadBalance

// Rebalancer increases weights on servers that perform better than others. It also rolls back to original weights
// if the servers have changed. It is designed as a wrapper on top of the roundrobin.
type Rebalancer struct {
	mutex sync.RWMutex
	// server records that remember original weights
	rbnodeMap         map[string]*RBNode // map[id]*RBNode
	balancer          LoadBalance
	isUpsertBalancer  bool
	isWithoutBreaker  bool
	cbreakerThreshold int // 健康节点阈值 低于此阈值禁止熔断
	succeedCondition  func(code int) bool
	newLoadBalance    NewLoadBalanceFunc
	newCBreaker       NewCBreakerFunc
	newRatio          NewRatioFunc
	log               *log.Logger
}

// RBNode record that keeps track of the original weight supplied by user
type RBNode struct {
	mutex      sync.RWMutex
	address    string
	origWeight int // original weight supplied by user
	curWeight  int // current weight
	node       interface{}
	cbreaker   *cbreaker
	ratio      *ratio
}

// Meter measures server performance and returns it's relative value via rating
type Meter interface {
	Rating() float64
	Record(int, time.Duration)
	IsReady() bool
}

// RebalancerOption - functional option setter for rebalancer
type RebalancerOption func(*Rebalancer) error

// RebalancerNew sets a reload builder function
func RebalancerNew(newLoadBalance NewLoadBalanceFunc) RebalancerOption {
	return func(r *Rebalancer) error {
		r.newLoadBalance = newLoadBalance
		return nil
	}
}

func RebalancerLogger(l *log.Logger) RebalancerOption {
	return func(rb *Rebalancer) error {
		rb.log = l
		return nil
	}
}

func RebalancerUpsert() RebalancerOption {
	return func(rb *Rebalancer) error {
		rb.isUpsertBalancer = true
		return nil
	}
}

func RebalancerSucceedCondition(condition func(int) bool) RebalancerOption {
	return func(rb *Rebalancer) error {
		rb.succeedCondition = condition
		return nil
	}
}

func RebalancerNoBreaker() RebalancerOption {
	return func(rb *Rebalancer) error {
		rb.isWithoutBreaker = true
		return nil
	}
}

func RebalancerIsWithoutBreaker(isWithoutBreaker bool) RebalancerOption {
	return func(rb *Rebalancer) error {
		rb.isWithoutBreaker = isWithoutBreaker
		return nil
	}
}

func RebalancerBreakerThreshold(threshold int) RebalancerOption {
	return func(rb *Rebalancer) error {
		rb.cbreakerThreshold = threshold
		return nil
	}
}

// NewRebalancer creates a new Rebalancer
func NewRebalancer(opts ...RebalancerOption) (*Rebalancer, error) {
	rb := &Rebalancer{
		rbnodeMap:         make(map[string]*RBNode),
		log:               log.StandardLogger(),
		cbreakerThreshold: defaultCBreakerThreshold,
	}

	for _, o := range opts {
		if err := o(rb); err != nil {
			return nil, err
		}
	}

	if rb.succeedCondition == nil {
		rb.succeedCondition = func(code int) bool {
			if code == 404 || (code >= 499 && code <= 600) {
				return false
			}
			return true
		}
	}

	rb.balancer = rb.newLoadBalance()

	if rb.newCBreaker == nil {
		rb.newCBreaker = func(options ...CBreakerOption) (*cbreaker, error) {
			if rb.isWithoutBreaker {
				return nil, nil
			}
			st := &Settings{
				OnStateChange:     rb.onStateChange,
				IsInterceptChange: rb.isInterceptChange,
			}
			for _, o := range options {
				o(st)
			}

			return NewCBreaker(st)
		}
	}

	if rb.newRatio == nil {
		rb.newRatio = func() *ratio {
			return NewRatio(defaultRecoveryDuration)
		}
	}

	return rb, nil
}

// adjustWeights 根据metrics算出节点的最新权重
// @王冀航 实现节点熔断
func (rb *Rebalancer) adjustWeight(node *RBNode, curWeight int) {
	if curWeight > node.origWeight {
		curWeight = node.origWeight
	}

	// 调整节点
	node.mutex.Lock()
	node.curWeight = curWeight
	node.mutex.Unlock()
	// 上报 调整loadbalance
	if rb.isUpsertBalancer {
		if err := rb.Upsert(node.address, curWeight, node.node); err != nil {
			rb.log.Errorf("更新[%s]权重为[%d]错误: %v", node.address, curWeight, err)
			return
		}
	} else {
		// 重建balancer
		balancer := rb.newLoadBalance()
		rb.mutex.RLock()
		for _, node := range rb.rbnodeMap {
			node.mutex.RLock()
			cweight := node.curWeight
			node.mutex.RUnlock()
			if cweight > 0 {
				if err := balancer.Upsert(node.address, cweight, node.node); err != nil {
					rb.log.Errorf("创建node[%s]权重[%d]错误: %v", node.address, node.curWeight, err)
				}
			}
		}
		rb.mutex.RUnlock()

		rb.mutex.Lock()
		rb.balancer = balancer
		rb.mutex.Unlock()
	}
}

// recordMetrics 记录metrics
// @王冀航 实现节点熔断
func (rb *Rebalancer) RecordMetrics(address string, code int) {
	success := rb.succeedCondition(code)

	rb.mutex.RLock()
	node := rb.rbnodeMap[address]
	rb.mutex.RUnlock()

	if node == nil {
		rb.log.Errorf("not find RbNode by address[%s]", address)
		return
	}

	// 如果是half_open，修改权重
	if node.cbreaker.State() == StateHalfOpen {
		if node.ratio == nil {
			node.ratio = rb.newRatio()
		}
		rb.adjustWeight(node, node.ratio.CalculateWeight(node.origWeight))
	}

	node.cbreaker.Record(success)
}

func (rb *Rebalancer) onStateChange(id string, from, to State) {
	rb.mutex.RLock()
	node := rb.rbnodeMap[id]
	rb.mutex.RUnlock()

	if node == nil {
		rb.log.Errorf("not found node by id[%s].", id)
		return
	}
	rb.log.Infof("id[%s] state from %s to %s", id, from, to)
	switch to {
	case StateOpen:
		// 调整权重为0
		rb.adjustWeight(node, 0)
		go func() {
			time.Sleep(defaultFallbackDuration)

			node.ratio = rb.newRatio()
			rb.adjustWeight(node, 1) // 放量
		}()
	case StateClosed:
		rb.adjustWeight(node, node.origWeight)
	}

}

// isInterceptCBreaker 是否拦截熔断
// case1：node < 2 true
// case2: (health + 1) / all > threshold_value true
// otherwise: false
func (rb *Rebalancer) isInterceptChange(name string, from, to State) bool {
	if to != StateOpen { // half-open，closed
		return false
	}

	// case 1
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	count := len(rb.rbnodeMap)
	if count < 2 {
		return true
	}

	// case 2
	health := -1 // 当前节点非正常
	for _, node := range rb.rbnodeMap {
		if node.cbreaker.state == StateClosed {
			health++
		}
	}
	// 健康节点比例 < 阈值
	if float64(health+1)/float64(count)*100 < float64(rb.cbreakerThreshold) {
		rb.log.Warnf("id[%s] state change is intercepted[from: %s to %s]", name, from, to)
		return true
	}

	return false
}

// Reload 节点变化通知
// nodes 服务发现变更的节点
// 根据nodes进行状态转移判断
// @陈川
func (rb *Rebalancer) Reload(serviceName string, nodes []*degrade.Node) error {
	// 通知节点变化
	doUpdateList := func(nodes []*degrade.Node, isPanic bool) {
		rbnodeMap := make(map[string]*RBNode)
		balancer := rb.newLoadBalance()
		lastNodes := rb.rbnodeMap
		for _, node := range nodes {
			var cb *cbreaker = nil
			if !isPanic {
				var err error
				cb, err = rb.newCBreaker(CBreakerWithAddress(node.Address))
				if err != nil {
					rb.log.Errorf("address[%s], weight[%d], err: %v", node.Address, node.Weight, err)
				}
			}

			rbnodeMap[node.Address] = &RBNode{
				address:    node.Address,
				origWeight: node.Weight,
				curWeight:  node.Weight,
				node:       node.Node,
				cbreaker:   cb,
				ratio:      rb.newRatio(),
			}

			if err := balancer.Upsert(node.Address, node.Weight, node.Node); err != nil {
				rb.log.Errorf("address[%s], weight[%d], err: %v", node.Address, node.Weight, err)
			}
		}

		rb.mutex.Lock()
		rb.rbnodeMap = rbnodeMap
		rb.balancer = balancer
		rb.mutex.Unlock()

		rb.mutex.RLock()
		for address := range lastNodes {
			degrade.EventDeleteLabelValues(degrade.EventBreakerOpenStatus, address)
			degrade.EventDeleteLabelValues(degrade.EventBreakerHalfOpenStatus, address)
		}
		rb.mutex.RUnlock()
	}

	// 触发状态转移，当发生错误，需要在外层传递数据
	return degrade.UpdateList(serviceName, nodes, doUpdateList)
	// doUpdateList(nodes,false) // 当需要剔除服务降级逻辑时调用
	// return nil
}

// DiscoveryToHttpNode http 数据转换示例
func (rb *Rebalancer) DiscoveryToHttpNode(list *qudiscovery.ServiceList) ([]*degrade.Node, error) {
	nodes := make([]*degrade.Node, 0, len(list.Services))
	for i := range list.Services {
		node := &degrade.Node{
			Address:     list.Services[i].ID,
			Name:        list.Services[i].Name,
			Host:        list.Services[i].Address,
			Port:        strconv.FormatInt(int64(list.Services[i].Port), 10),
			Weight:      list.Services[i].Weight,
			HealthCheck: list.Services[i].HealthCheck,
		}

		url, err := url.Parse("http://" + list.Services[i].Address + ":" + strconv.FormatInt(int64(list.Services[i].Port), 10))
		if err != nil {
			rb.log.Errorf("get node.url error: %v [%#v]", err, list.Services[i])
			return nil, err
		}

		node.Node = &degrade.HttpNextNode{
			ID:  list.Services[i].ID,
			URL: url,
		}

		if len(list.Services[i].Meta) > 0 {
			node.IsOffline = list.Services[i].Meta[degrade.MetaKeyStatus] == degrade.MetaStatusOffline
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// DiscoveryToRpcNode rpc 数据转换示例
func (rb *Rebalancer) DiscoveryToRpcNode(list *qudiscovery.ServiceList) []*degrade.Node {
	nodes := make([]*degrade.Node, 0, len(list.Services))
	for _, service := range list.Services {
		node := &degrade.Node{
			Address:     service.ID,
			Name:        service.Name,
			Host:        service.Address,
			Port:        strconv.FormatInt(int64(service.Port), 10),
			Weight:      service.Weight,
			HealthCheck: service.HealthCheck,
			Node: &degrade.RpcNextNode{
				Id:      service.ID,
				Address: service.Address + ":" + strconv.FormatInt(int64(service.Port), 10),
			},
		}

		if len(service.Meta) > 0 {
			node.IsOffline = service.Meta[degrade.MetaKeyStatus] == degrade.MetaStatusOffline
		}
		nodes = append(nodes, node)
	}

	return nodes
}

// CloseDegrade 关闭服务降级
func (rb *Rebalancer) CloseDegrade() {
	degrade.CloseDegrade()

}

func (rb *Rebalancer) upsert(id string, weight int, node interface{}) error {
	return nil
}

// Upsert 插入节点
func (rb *Rebalancer) Upsert(id string, weight int, node interface{}) error {
	cb, err := rb.newCBreaker(CBreakerWithAddress(id))
	if err != nil {
		return err
	}
	rbNode := &RBNode{
		address:    id,
		origWeight: weight,
		curWeight:  weight,
		node:       node,
		cbreaker:   cb,
		ratio:      rb.newRatio(),
	}

	rb.mutex.Lock()
	rb.rbnodeMap[id] = rbNode
	degrade.EventDeleteLabelValues(degrade.EventBreakerOpenStatus, id)
	degrade.EventDeleteLabelValues(degrade.EventBreakerHalfOpenStatus, id)
	rb.mutex.Unlock()

	rb.mutex.RLock()
	rb.balancer.Upsert(id, weight, node)
	rb.mutex.RUnlock()

	return nil
}

// UpsertNotBreaker 去除节点熔断地插入节点
func (rb *Rebalancer) UpsertNotBreaker(id string, weight int, node interface{}) {
	rbNode := &RBNode{
		address:    id,
		origWeight: weight,
		curWeight:  weight,
		node:       node,
		cbreaker:   nil,
		ratio:      rb.newRatio(),
	}

	rb.mutex.Lock()
	rb.rbnodeMap[id] = rbNode
	rb.mutex.Unlock()

	rb.mutex.RLock()
	rb.balancer.Upsert(id, weight, node)
	rb.mutex.RUnlock()
}

// Next 选择节点
func (rb *Rebalancer) Next() (interface{}, error) {
	rb.mutex.RLock()
	balancer := rb.balancer
	rb.mutex.RUnlock()

	return balancer.Next()
}
