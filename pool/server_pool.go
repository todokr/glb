package pool

import (
	"todokr.github.io/glb/backend"
)

// ServerPool is a collection of targets
type ServerPool struct {
	targets []*backend.Target
	Chooser
}

type Chooser interface {
	Choose() *backend.Target
}
