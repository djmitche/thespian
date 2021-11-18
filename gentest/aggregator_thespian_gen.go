// code generaged by thespian; DO NOT EDIT

package gentest

import "github.com/djmitche/thespian"

// --- Aggregator

// Aggregator is the public handle for aggregator actors.
type Aggregator struct {
	stopChan      chan<- struct{}
	incrementChan chan<- string
	incrSender    StringSender
}

// Stop sends a message to stop the actor.
func (a *Aggregator) Stop() {
	a.stopChan <- struct{}{}
}

// Increment sends the Increment message to the actor.
func (a *Aggregator) Increment(m string) {
	a.incrementChan <- m
}

// Incr sends to the actor's Incr mailbox.
func (a *Aggregator) Incr(m string) {
	// TODO: generate this based on the mbox kind
	a.incrSender.C <- m
}

// --- aggregator

func (a aggregator) spawn(rt *thespian.Runtime) *Aggregator {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	// TODO: generate based on mbox kind
	incrMailbox := NewStringMailbox()
	// TODO: generate based on mbox kind
	a.incrReceiver = incrMailbox.Receiver()

	handle := &Aggregator{
		stopChan:      a.StopChan,
		incrementChan: a.incrementChan,
		// TODO: generate based on mbox kind
		incrSender: incrMailbox.Sender(),
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
		case m := <-a.incrementChan:
			a.handleIncrement(m)
		case m := <-*a.flushTimer.C:
			a.handleFlush(m)
		// TODO: generate this based on the mbox kind
		case m := <-a.incrReceiver.C:
			a.handleIncr(m)
		}
	}
}

func (a *aggregator) cleanup() {
	a.flushTimer.Stop()
	// TODO: clean up mboxes too
	a.Runtime.ActorStopped(&a.ActorBase)
}
