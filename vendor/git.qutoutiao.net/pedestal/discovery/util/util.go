package util

import (
	"context"
	"net"
	"time"

	"golang.org/x/exp/rand"

	"git.qutoutiao.net/pedestal/discovery/registry"
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

func SlidingDuration(d time.Duration) time.Duration {
	sliding := rand.Int63() % (int64(d) / 16)

	return d + time.Duration(sliding)
}

// ShuffleService sort services with random.
func ShuffleService(services []*registry.Service) {
	if len(services) < 2 {
		return
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	rand.Shuffle(len(services), func(i, j int) {
		services[i], services[j] = services[j], services[i]
	})
}

// ResolveRandomHost returns a random host from DNS query result.
func ResolveRandomHost(hostname string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r := net.Resolver{}

	addrs, err := r.LookupHost(ctx, hostname)
	if err != nil {
		return "", err
	}

	idx := 0
	if len(addrs) > 1 {
		idx = int(rand.Int31()) % len(addrs)
	}

	return addrs[idx], nil
}
