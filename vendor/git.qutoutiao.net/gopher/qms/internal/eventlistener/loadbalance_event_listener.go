package eventlistener

// constants for loadbalancer strategy name, and timeout
const (
//LoadBalanceKey is variable of type string that matches load balancing events
//LoadBalanceKey          = "^qms\\.loadbalance\\."
//regex4normalloadbalance = "^qms\\.loadbalance\\.(strategy|SessionStickinessRule|retryEnabled|retryOnNext|retryOnSame|backoff)"
)

//LoadbalancingEventListener is a struct
type LoadbalancingEventListener struct {
	Key string
}

//Event is a method used to handle a load balancing event
//func (e *LoadbalancingEventListener) Event(evt *event.Event) {
//	qlog.Tracef("LB event, key: %s, type: %s", evt.Key, evt.EventType)
//	if err := config.ReadLBFromArchaius(); err != nil {
//		qlog.Error("can not unmarshal new lb config: " + err.Error())
//	}
//	archaius.SaveToLBCache(config.GetLoadBalancing())
//}
