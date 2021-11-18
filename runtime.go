package thespian

import "sync/atomic"

// A Thespian runtime manages a collection of actors.
type Runtime struct {
	// lastID is atomically incremented to generate actor IDs
	lastID uint64
}

func NewRuntime() *Runtime {
	return &Runtime{
		lastID: 0,
	}
}

func (rt *Runtime) Register(base *ActorBase) {
	if base.ID != 0 {
		panic("ActorBase has already been registered")
	}
	base.ID = atomic.AddUint64(&rt.lastID, 1)
	base.Runtime = rt
	base.StopChan = make(chan struct{}, 5)
}
