package pool

import (
	"testing"

	"todokr.github.io/glb/backend"
)

func TestRoundRobinServerPool_nextIndex(t *testing.T) {
	pool := NewRoundRobinServerPool([]*backend.Target{{}, {}, {}})
	if pool.current != 0 {
		t.Error("Expected 0")
	}
	if pool.nextIndex() != 1 {
		t.Error("Expected 1")
	}
	if pool.nextIndex() != 2 {
		t.Error("Expected 2")
	}
	if pool.nextIndex() != 0 {
		t.Error("Expected 0")
	}
}
