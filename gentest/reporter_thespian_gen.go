// code generaged by thespian; DO NOT EDIT

package gentest

import "github.com/djmitche/thespian"

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
