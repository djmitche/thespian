package super

import (
	"testing"
	"time"

	"github.com/djmitche/thespian"
	"github.com/stretchr/testify/require"
)

type supervisor struct {
	supervisorBase
	childID uint64
	events  chan thespian.SuperEvent
}

func (a *supervisor) handleStart() {
	a.rx.supervise(a.childID)
}

func (a *supervisor) handleSuperEvent(ev thespian.SuperEvent) {
	a.events <- ev
}

type supervisee struct {
	superviseeBase
	beUnhealthy bool
}

func (a *supervisee) handleStart() {
	if a.beUnhealthy {
		time.Sleep(5 * time.Second)
	}
}

func TestStoppedEvent(t *testing.T) {
	rt := thespian.NewRuntime()
	ch := make(chan thespian.SuperEvent, 10)
	child := SuperviseeBuilder{supervisee{beUnhealthy: false}}.spawn(rt)
	parent := SupervisorBuilder{supervisor{childID: child.ID, events: ch}}.spawn(rt)

	var ev thespian.SuperEvent
	go func() {
		// TODO: race here between parent supervising and child stopping
		child.Stop()
		select {
		case ev = <-ch:
		case <-time.After(1 * time.Second):
		}
		parent.Stop()
	}()

	rt.Run()

	require.Equal(t, thespian.SuperEvent{
		Event: thespian.StoppedActor,
		ID:    child.ID,
	}, ev)
}
