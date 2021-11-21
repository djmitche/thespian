// code generated by thespian; DO NOT EDIT

package ticker

import (
	"github.com/djmitche/thespian"
	import1 "github.com/djmitche/thespian/mailbox"
)

// TimerBuilder is used to buidl new Timer actors.
type TimerBuilder struct {
	timer
}

func (bldr TimerBuilder) spawn(rt *thespian.Runtime) *TimerTx {
	reg := rt.Register()

	rx := &TimerRx{
		id:         reg.ID,
		rt:         rt,
		stopChan:   reg.StopChan,
		superChan:  reg.SuperChan,
		healthChan: reg.HealthChan,
		tkr:        import1.NewTickerRx(rt),
	}

	tx := &TimerTx{
		ID:       reg.ID,
		stopChan: reg.StopChan,
	}

	// copy to a new timer instance
	pvt := bldr.timer
	pvt.rt = rt
	pvt.rx = rx
	pvt.tx = tx

	go pvt.loop()
	return tx
}

// TimerRx contains the Rx sides of the mailboxes, for access from the
// Timer implementation.
type TimerRx struct {
	id uint64
	rt *thespian.Runtime

	stopChan   <-chan struct{}
	superChan  <-chan thespian.SuperEvent
	healthChan <-chan struct{}
	tkr        import1.TickerRx
}

// supervise starts supervision of the actor identified by otherID.
// It is a shortcut to thespian.Runtime.Supervize.
func (rx *TimerRx) supervise(otherID uint64) {
	rx.rt.Supervise(rx.id, otherID)
}

// unsupervise stops supervision of the actor identified by otherID.
// It is a shortcut to thespian.Runtime.Unupervize.
func (rx *TimerRx) unsupervise(otherID uint64) {
	rx.rt.Unsupervise(rx.id, otherID)
}

// TimerTx is the public handle for Timer actors.
type TimerTx struct {
	// ID is the unique ID of this actor
	ID       uint64
	stopChan chan<- struct{}
}

// Stop sends a message to stop the actor.  This does not wait until
// the actor has stopped.
func (a *TimerTx) Stop() {
	select {
	case a.stopChan <- struct{}{}:
	default:
	}
}

func (a *timer) loop() {
	rx := a.rx
	defer func() {
		rx.tkr.Stop()
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
		case t := <-rx.tkr.Chan():
			a.handleTkr(t)
		}
	}
}