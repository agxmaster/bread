package balancer

const (
	RoundRobin         Algorithm = "roundrobin"
	WeightedRoundRobin Algorithm = "weightedroundrobin"
	Random             Algorithm = "random"
	WeightedRandom     Algorithm = "weightedrandom"
)

type Algorithm string

func (alg Algorithm) IsValid() bool {
	switch alg {
	case RoundRobin, WeightedRoundRobin, Random, WeightedRandom:
		return true
	}

	return false
}

func (alg Algorithm) String() string {
	return string(alg)
}
