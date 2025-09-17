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

func (b *balancer) index() (res int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	res = b.idx
	
	if b.idx == b.max-1 {
		b.idx = 0
	} else {
		b.idx++
	}

	return
}
