package wrr

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// RoundRandWeight 通用按权随机负载均衡
type RoundRandWeight struct {
	mutex sync.RWMutex

	servers    []*server
	serversMap map[string]*server
	serverList atomic.Value // *serverList
}

// NewRoundRandWeight 创建 NewRoundRandWeight
func NewRoundRandWeight() *RoundRandWeight {
	return &RoundRandWeight{
		servers:    []*server{},
		serversMap: make(map[string]*server),
	}
}

// Next 获取下一个节点
func (r *RoundRandWeight) Next() (interface{}, error) {
	s, _ := r.serverList.Load().(*serverList)
	if s == nil {
		return nil, fmt.Errorf("no endpoints in the pool")
	}
	u := s.nextServer()
	if u == nil {
		return nil, fmt.Errorf("no endpoints in the pool")
	}
	return u, nil
}

// Upsert 添加节点
func (r *RoundRandWeight) Upsert(id string, weight int, u interface{}) error {
	if u == nil {
		return fmt.Errorf("server URL can't be nil")
	}
	srv := &server{val: u}
	srv.weight = weight
	if srv.weight == 0 {
		return fmt.Errorf("weight can not be zero")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if s := r.findServer(id); s != nil {
		r.resetState()
		return nil
	}

	srv.index = len(r.servers)
	r.servers = append(r.servers, srv)
	r.serversMap[id] = srv
	r.resetState()
	return nil
}

// Remove 删除节点
func (r *RoundRandWeight) Remove(id string) error {
	r.mutex.Lock()

	e := r.findServer(id)
	if e == nil {
		r.mutex.Unlock()
		return fmt.Errorf("server not found")
	}
	for i := e.index + 1; i < len(r.servers); i++ {
		r.servers[i].index--
	}
	r.servers = append(r.servers[:e.index], r.servers[e.index+1:]...)
	delete(r.serversMap, id)
	r.resetState()
	r.mutex.Unlock()
	return nil
}

// Servers 获取所有节点
func (r *RoundRandWeight) Servers() []interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	out := make([]interface{}, len(r.servers))
	for i, srv := range r.servers {
		out[i] = srv.val
	}
	return out
}

func (r *RoundRandWeight) findServer(id string) *server {
	return r.serversMap[id]
}

func (r *RoundRandWeight) resetState() {
	s := buildServerList(r.servers)
	r.serverList.Store(s)
}

// Set additional parameters for the server can be supplied when adding server
type server struct {
	val interface{}
	// Relative weight for the enpoint to other enpoints in the load balancer
	weight int

	index int // 数组下表
}

// serverList 服务列表选择
type serverList struct {
	servers []*server
	r       *rand.Rand
}

func buildServerList(servers []*server) *serverList {
	s := &serverList{
		r: rand.New(newLockedSource()),
	}
	if len(servers) < 0 {
		return s
	}
	s.servers = make([]*server, 0, len(servers))
	for _, v := range servers {
		if v.weight <= 0 {
			continue
		}
		tmp := *v
		s.servers = append(s.servers, &tmp)
	}
	for i := 1; i < len(servers); i++ {
		s.servers[i].weight += s.servers[i-1].weight
	}
	return s
}

func (s *serverList) nextServer() interface{} {
	n := len(s.servers)
	if n == 1 {
		return s.servers[0].val
	}
	if n == 0 {
		return nil
	}
	val := s.r.Intn(s.servers[n-1].weight)
	li, ri := 0, n
	for li < ri {
		m := (li + ri) >> 1
		if s.servers[m].weight <= val {
			li = m + 1
		} else if s.servers[m].weight > val {
			ri = m
		}
	}
	return s.servers[li].val
}

type lockedSource struct {
	lk  sync.Mutex
	src rand.Source
}

func newLockedSource() rand.Source {
	return &lockedSource{
		src: rand.NewSource(time.Now().UnixNano()),
	}
}

func (r *lockedSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}
