// code generaged by thespian; DO NOT EDIT

package gentest

// TODO: use variable
import "github.com/djmitche/thespian"

// --- Aggregator

// Aggregator is the public handle for aggregator actors.
type Aggregator struct {
	stopChan chan<- struct{}
	// TODO: generate this based on the mbox kind
	incrSender StringSender
	// TODO: generate this based on the mbox kind
}

// Stop sends a message to stop the actor.
func (a *Aggregator) Stop() {
	a.stopChan <- struct{}{}
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
	// TODO: generate based on mbox kind
	a.incrReceiver = incrMailbox.Receiver()
	// TODO: generate based on mbox kind
	a.flushReceiver = NewTickerReceiver()

	handle := &Aggregator{
		stopChan: a.StopChan,
		// TODO: generate based on mbox kind
		incrSender: incrMailbox.Sender(),
		// TODO: generate based on mbox kind
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
		case m := <-a.incrReceiver.C:
			a.handleIncr(m)
		// TODO: generate this based on the mbox kind
		case t := <-a.flushReceiver.Chan():
			a.handleFlush(t)
		}
	}
}

func (a *aggregator) cleanup() {
	// TODO: clean up mboxes too
	a.Runtime.ActorStopped(&a.ActorBase)
}
