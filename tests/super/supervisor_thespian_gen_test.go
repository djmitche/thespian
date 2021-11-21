// code generated by thespian; DO NOT EDIT

package super

import (
	"github.com/djmitche/thespian"
)

// SupervisorBuilder is used to buidl new Supervisor actors.
type SupervisorBuilder struct {
	supervisor
}

func (bldr SupervisorBuilder) spawn(rt *thespian.Runtime) *SupervisorTx {
	reg := rt.Register()

	rx := &SupervisorRx{
		id:         reg.ID,
		rt:         rt,
		stopChan:   reg.StopChan,
		superChan:  reg.SuperChan,
		healthChan: reg.HealthChan,
	}

	tx := &SupervisorTx{
		ID:       reg.ID,
		stopChan: reg.StopChan,
	}

	// copy to a new supervisor instance
	pvt := bldr.supervisor
	pvt.rt = rt
	pvt.rx = rx
	pvt.tx = tx

	go pvt.loop()
	return tx
}

// SupervisorRx contains the Rx sides of the mailboxes, for access from the
// Supervisor implementation.
type SupervisorRx struct {
	id uint64
	rt *thespian.Runtime

	stopChan   <-chan struct{}
	superChan  <-chan thespian.SuperEvent
	healthChan <-chan struct{}
}

// supervise starts supervision of the actor identified by otherID.
// It is a shortcut to thespian.Runtime.Supervize.
func (rx *SupervisorRx) supervise(otherID uint64) {
	rx.rt.Supervise(rx.id, otherID)
}

// unsupervise stops supervision of the actor identified by otherID.
// It is a shortcut to thespian.Runtime.Unupervize.
func (rx *SupervisorRx) unsupervise(otherID uint64) {
	rx.rt.Unsupervise(rx.id, otherID)
}

// SupervisorTx is the public handle for Supervisor actors.
type SupervisorTx struct {
	// ID is the unique ID of this actor
	ID       uint64
	stopChan chan<- struct{}
}

// Stop sends a message to stop the actor.  This does not wait until
// the actor has stopped.
func (a *SupervisorTx) Stop() {
	select {
	case a.stopChan <- struct{}{}:
	default:
	}
}

func (a *supervisor) loop() {
	rx := a.rx
	defer func() {
		a.rt.ActorStopped(a.rx.id)
	}()
	a.handleStart()
	for {
		select {
		case <-rx.healthChan:
			// nothing to do
		case ev := <-rx.superChan:
			a.handleSuperEvent(ev)
		case <-rx.stopChan:
			a.handleStop()
			return
		}
	}
}
