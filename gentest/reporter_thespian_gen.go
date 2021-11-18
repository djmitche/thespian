// code generaged by thespian; DO NOT EDIT

package gentest

// TODO: use variable
import "github.com/djmitche/thespian"

// --- Reporter

// Reporter is the public handle for reporter actors.
type Reporter struct {
	stopChan chan<- struct{}
	// TODO: generate this based on the mbox kind
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

// Report sends to the actor's Report mailbox.
func (a *Reporter) Report(m []string) {
	// TODO: generate this based on the mbox kind
	a.reportTx.C <- m
}

// --- reporter

func (a reporter) spawn(rt *thespian.Runtime) *Reporter {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	// TODO: generate based on mbox kind
	reportMailbox := NewStringSliceMailbox()
	// TODO: generate based on mbox kind
	a.reportRx = reportMailbox.Rx()

	handle := &Reporter{
		stopChan: a.StopChan,
		// TODO: generate based on mbox kind
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
		// TODO: generate this based on the mbox kind
		case m := <-a.reportRx.C:
			a.handleReport(m)
		}
	}
}

func (a *reporter) cleanup() {
	// TODO: clean up mboxes too
	a.Runtime.ActorStopped(&a.ActorBase)
}
