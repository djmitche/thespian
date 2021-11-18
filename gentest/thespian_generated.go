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

// --- Reporter

// Reporter is the public handle for reporter actors.
type Reporter struct {
	stopChan   chan<- struct{}
	reportChan chan<- []string
}

// Stop sends a message to stop the actor.
func (a *Reporter) Stop() {
	a.stopChan <- struct{}{}
}

// Report sends the Report message to the actor.
func (a *Reporter) Report(m []string) {
	a.reportChan <- m
}

// --- reporter

func (a reporter) spawn(rt *thespian.Runtime) *Reporter {
	rt.Register(&a.ActorBase)
	handle := &Reporter{
		stopChan:   a.StopChan,
		reportChan: a.reportChan,
	}
	go a.loop()
	return handle
}

func (a *reporter) loop() {
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
		case m := <-a.reportChan:
			a.handleReport(m)
		}
	}
}

func (a *reporter) cleanup() {
	a.Runtime.ActorStopped(&a.ActorBase)
}
