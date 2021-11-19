// code generaged by thespian; DO NOT EDIT

package gentest

import "github.com/djmitche/thespian"

// --- Reporter

// Reporter is the public handle for reporter actors.
type Reporter struct {
	stopChan chan<- struct{}
	reportTx StringSliceTx
}

// Stop sends a message to stop the actor.  This does not wait until
// the actor has stopped.
func (a *Reporter) Stop() {
	select {
	case a.stopChan <- struct{}{}:
	default:
	}
}

// Report sends to the actor's report mailbox.
func (a *Reporter) Report(m []string) {
	a.reportTx.C <- m
}

// --- reporter

func (a reporter) spawn(rt *thespian.Runtime) *Reporter {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	reportMailbox := NewStringSliceMailbox()
	a.reportRx = reportMailbox.Rx()

	handle := &Reporter{
		stopChan: a.StopChan,
		reportTx: reportMailbox.Tx(),
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
		case m := <-a.reportRx.C:
			a.handleReport(m)
		}
	}
}

func (a *reporter) cleanup() {

	a.Runtime.ActorStopped(&a.ActorBase)
}
