package thespian

// TODO: it'd be nice to have these fields be private
type ActorBase struct {
	ID       uint64
	Runtime  *Runtime
	StopChan chan struct{}
}

// --- default implementations

func (a *ActorBase) HandleStart() error {
	// could be overridden by user impl
	return nil
}

func (a *ActorBase) HandleStop() error {
	// could be overridden by user impl
	return nil
}
