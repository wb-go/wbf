package dbpg

import "sync"

type balancer struct {
	idx int
	max int // Кол-во slaves

	mu *sync.Mutex
}

func newBalancer(max int) *balancer {
	return &balancer{
		idx: 0,
		max: max,
		mu:  new(sync.Mutex),
	}
}

func (b *balancer) index() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.max <= 0 {
		return 0
	}

	res := b.idx
	b.idx = (b.idx + 1) % b.max
	
	return res
}
