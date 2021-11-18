// code generaged by thespian; DO NOT EDIT

package gentest

import "github.com/djmitche/thespian"

// --- Aggregator

// Aggregator is the public handle for aggregator actors.
type Aggregator struct {
	stopChan      chan<- struct{}
	incrementChan chan<- string
}

// Stop sends a message to stop the actor.
func (a *Aggregator) Stop() {
	a.stopChan <- struct{}{}
}

// Increment sends the Increment message to the actor.
func (a *Aggregator) Increment(m string) {
	a.incrementChan <- m
}

// --- aggregator

func (a aggregator) spawn(rt *thespian.Runtime) *Aggregator {
	rt.Register(&a.ActorBase)
	handle := &Aggregator{
		stopChan:      a.StopChan,
		incrementChan: a.incrementChan,
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
		}
	}
}

func (a *aggregator) cleanup() {
	a.flushTimer.Stop()
	a.Runtime.ActorStopped(&a.ActorBase)
}