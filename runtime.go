package thespian

import "time"

// TODO: it'd be nice to have these fields be private
type ActorBase struct {
	ID       uint64
	StopChan chan struct{}
}

func NewActorBase() ActorBase {
	return ActorBase{
		ID:       13, // TODO
		StopChan: make(chan struct{}, 5),
	}
}

// --- Timer

type Timer struct {
	C      *<-chan time.Time
	ticker *time.Ticker
}

func (t *Timer) Tick(dur time.Duration) {
	t.ticker = time.NewTicker(dur)
	t.C = &t.ticker.C
}

func (t *Timer) Stop() {
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
	t.C = nil
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
