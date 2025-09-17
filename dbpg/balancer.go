package dbpg

import "sync"

type balancer struct {
	idx     int
	maxSize int // Кол-во slaves

	mu *sync.Mutex
}

func newBalancer(maxSize int) *balancer {
	return &balancer{
		idx:     0,
		maxSize: maxSize,
		mu:      new(sync.Mutex),
	}
}

func (b *balancer) index() (res int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	res = b.idx

	if b.idx == b.maxSize-1 {
		b.idx = 0
	} else {
		b.idx++
	}

	return
}
