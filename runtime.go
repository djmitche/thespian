package thespian

import "time"

type AgentBase struct {
	id       uint64
	stopChan chan struct{}
}

func NewAgentBase() AgentBase {
	return AgentBase{
		id:       13, // TODO
		stopChan: make(chan struct{}, 5),
	}
}

// --- Timer

type Timer struct {
	c      *<-chan time.Time
	ticker *time.Ticker
}

func (t *Timer) Tick(dur time.Duration) {
	t.ticker = time.NewTicker(dur)
	t.c = &t.ticker.C
}

func (t *Timer) Stop() {
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
	t.c = nil
}

// --- default implementations

func (a *AgentBase) handleStart() error {
	// could be overridden by user impl
	return nil
}

func (a *AgentBase) handleStop() error {
	// could be overridden by user impl
	return nil
}
