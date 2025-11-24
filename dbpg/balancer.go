package dbpg

import "sync"

type balancer struct {
	idx       int
	maxSlaves int // Number of slave connections.

	mu *sync.Mutex
}

func newBalancer(maxSlaves int) *balancer {
	return &balancer{
		idx:       0,
		maxSlaves: maxSlaves,
		mu:        &sync.Mutex{},
	}
}

func (b *balancer) index() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.maxSlaves <= 0 {
		return 0
	}

	res := b.idx
	b.idx = (b.idx + 1) % b.maxSlaves

	return res
}
