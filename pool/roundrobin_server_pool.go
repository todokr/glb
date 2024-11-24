package pool

import (
	"sync/atomic"

	"todokr.github.io/glb/backend"
)

type RoundRobinServerPool struct {
	ServerPool
	current uint64
}

func (s *RoundRobinServerPool) nextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.targets)))
}

func (s *RoundRobinServerPool) Choose() *backend.Target {
	ni := s.nextIndex()
	l := len(s.targets) + ni
	for i := ni; i < l; i++ {
		idx := i % len(s.targets)
		candidate := s.targets[idx]
		if candidate.IsHealthy() {
			if idx != ni {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return candidate
		}
	}
	return nil
}

func NewRoundRobinServerPool(ts []*backend.Target) *RoundRobinServerPool {
	return &RoundRobinServerPool{
		ServerPool: ServerPool{targets: ts},
		current:    0,
	}
}
