// code generaged by thespian; DO NOT EDIT

package gentest

// TODO: use variable
import "github.com/djmitche/thespian"

// --- Reporter

// Reporter is the public handle for reporter actors.
type Reporter struct {
	stopChan chan<- struct{}
	// TODO: generate this based on the mbox kind
	reportSender StringSliceSender
}

// Stop sends a message to stop the actor.
func (a *Reporter) Stop() {
	a.stopChan <- struct{}{}
}

// Report sends to the actor's Report mailbox.
func (a *Reporter) Report(m []string) {
	// TODO: generate this based on the mbox kind
	a.reportSender.C <- m
}

// --- reporter

func (a reporter) spawn(rt *thespian.Runtime) *Reporter {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	// TODO: generate based on mbox kind
	reportMailbox := NewStringSliceMailbox()
	// TODO: generate based on mbox kind
	a.reportReceiver = reportMailbox.Receiver()

	handle := &Reporter{
		stopChan: a.StopChan,
		// TODO: generate based on mbox kind
		reportSender: reportMailbox.Sender(),
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
		case m := <-a.reportReceiver.C:
			a.handleReport(m)
		}
	}
}

func (a *reporter) cleanup() {
	// TODO: clean up mboxes too
	a.Runtime.ActorStopped(&a.ActorBase)
}
