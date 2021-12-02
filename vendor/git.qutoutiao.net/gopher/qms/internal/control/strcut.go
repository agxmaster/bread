package control

//LoadBalancingConfig is a standardized model
type LoadBalancingConfig struct {
	Enabled      bool
	Strategy     string
	Filters      []string
	RetryEnabled bool
	RetryOnSame  int
	RetryOnNext  int
	BackOffKind  string
	BackOffMin   int
	BackOffMax   int

	SessionTimeoutInSeconds int
	SuccessiveFailedTimes   int
}

//RateLimitingConfig is a standardized model
type RateLimitingConfig struct {
	Enabled bool
	Key     string
	Rate    int
}

//EgressConfig is a standardized model
type EgressConfig struct {
	Hosts []string
	Ports []*EgressPort
}

//EgressPort protocol and the corresponding port
type EgressPort struct {
	Port     int32
	Protocol string
}
