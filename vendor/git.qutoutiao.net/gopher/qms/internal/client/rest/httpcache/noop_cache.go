package httpcache

type noopCache struct{}

func (n *noopCache) Get(key string) (responseBytes []byte, ok bool) {
	return
}

func (n *noopCache) Set(key string, responseBytes []byte) {}

func (n *noopCache) Delete(key string) {}
