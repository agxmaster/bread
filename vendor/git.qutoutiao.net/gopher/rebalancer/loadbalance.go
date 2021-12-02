package rebalancer

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"git.qutoutiao.net/gopher/qudiscovery"
	"git.qutoutiao.net/gopher/rebalancer/wrr"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
)

var (
	lbstore sync.Map
	lbmu    sync.RWMutex
)

// NewCallWrapper LoadBalance and returns a Call Wrapper
func NewCallWrapper(dis qudiscovery.Discovery) client.CallWrapper {
	return func(next client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			rb, err := getReBalancer(req.Service(), dis)
			if err != nil {
				fmt.Printf("err1 %+v", err)
				return err
			}

			nodeTmp, err := rb.Next()
			if err != nil {
				return fmt.Errorf("get node:%s err:%v", req.Service(), err)
			}
			node = nodeTmp.(*registry.Node)

			err = next(ctx, node, req, rsp, opts)
			if err != nil {
				return fmt.Errorf("run next wrapper err :%s err:%v", req.Service(), err)
			}

			// 注意：此处仅做示例
			type ResponseKey struct{}

			type Response struct {
				Code         int
				Body         []byte
				ReqBody      []byte
				HTTPResponse *http.Response
				HTTPRequest  *http.Request
				DialFailed   bool
			}
			resp := ctx.Value(ResponseKey{}).(*Response)

			// 记录节点的状态
			rb.RecordMetrics(node.Id, resp.Code)
			// 调整权重
			//rb.adjustWeights()
			return nil
		}
	}
}

func getReBalancer(serviceName string, dis qudiscovery.Discovery) (*Rebalancer, error) {
	var err error
	rb, exist := lbstore.Load(serviceName)
	if !exist {
		lbmu.Lock()
		defer lbmu.Unlock()
		rb, exist = lbstore.Load(serviceName)
		if !exist {
			rb, err = newRebalancerByDiscovery(serviceName, dis)
			if err != nil {
				return nil, err
			}
			lbstore.Store(serviceName, rb)
		}
	}
	return rb.(*Rebalancer), nil
}

func newRebalancerByDiscovery(serviceName string, dis qudiscovery.Discovery) (*Rebalancer, error) {
	// 创建一个空的rb对象
	rb, err := newReloadBalance([]*qudiscovery.Service{})
	if err != nil {
		return nil, err
	}

	checkChange, err := qudiscovery.NewCheckChange(func(list *qudiscovery.ServiceList) (interface{}, error) {
		fmt.Println("abc", list)
		nodes := rb.DiscoveryToRpcNode(list)

		// 更新操作
		if err = rb.Reload(serviceName, nodes); err != nil {
			rb.log.Error("newRebalancerByDiscovery.Reload error: ", err.Error())

			rb, err = newReloadBalance(list.Services)
			if err != nil {
				rb.log.Errorf("newRebalancerByDiscovery.newReloadBalance error: %v [%#v]", err, list.Services)
			}
			return nil, err
		}

		fmt.Println("testxy", nodes)

		// 另外一种方式, 重建rb
		return rb, nil
	})
	if err != nil {
		return nil, fmt.Errorf("service:%s NewCheckChange err:%v", serviceName, err)
	}

	serviceList, err := dis.GetServers(serviceName)
	if err != nil {
		return nil, fmt.Errorf("service:%s GetServers err:%v", serviceName, err)
	}

	val, _, err := checkChange.GetValue(serviceList)
	if err != nil {
		return nil, fmt.Errorf("service:%s checkChange.GetValue err:%v", serviceName, err)
	}
	rb = val.(*Rebalancer)

	return rb, nil
}

func newReloadBalance(nodes []*qudiscovery.Service) (*Rebalancer, error) {
	lb, err := NewRebalancer(RebalancerNew(func() LoadBalance {
		return wrr.NewRoundRandWeight()
	}))

	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if err := lb.Upsert(node.ID, node.Weight, &registry.Node{
			Id:      node.ID,
			Address: node.Address + ":" + strconv.Itoa(node.Port),
		}); err != nil {
			return nil, err
		}
	}
	return lb, nil
}
