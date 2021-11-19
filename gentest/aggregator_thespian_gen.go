// code generaged by thespian; DO NOT EDIT

package gentest

import "github.com/djmitche/thespian"

// --- Aggregator

// Aggregator is the public handle for aggregator actors.
type Aggregator struct {
	stopChan chan<- struct{}
	// TODO: generate this based on the mbox kind
	// TODO: generate this based on the mbox kind
	incrTx StringTx
}

// Stop sends a message to stop the actor.  This does not wait until
// the actor has stopped.
func (a *Aggregator) Stop() {
	select {
	case a.stopChan <- struct{}{}:
	default:
	}
}

// Incr sends to the actor's incr mailbox.
func (a *Aggregator) Incr(m string) {
	// TODO: generate this based on the mbox kind
	a.incrTx.C <- m
}

// --- aggregator

func (a aggregator) spawn(rt *thespian.Runtime) *Aggregator {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	incrMailbox := NewStringMailbox()
	a.flushRx = NewTickerRx()
	a.incrRx = incrMailbox.Rx()

	handle := &Aggregator{
		stopChan: a.StopChan,
		incrTx:   incrMailbox.Tx(),
	}
	go a.loop()
	return handle
}

func (a *aggregator) loop() {
	defer func() {
		a.cleanup()
	}()
	a.HandleStart()
	for {
		select {
		case <-a.HealthChan:
			// nothing to do
		case <-a.StopChan:
			a.HandleStop()
			return
		// TODO: generate this based on the mbox kind
		case t := <-a.flushRx.Chan():
			a.handleFlush(t)
		// TODO: generate this based on the mbox kind
		case m := <-a.incrRx.C:
			a.handleIncr(m)
		}
	}
}

func (a *aggregator) cleanup() {
	// TODO: clean up mboxes too
	a.Runtime.ActorStopped(&a.ActorBase)
}
