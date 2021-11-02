package thespian

// upper-case name is the external handle
type Reporter struct {
	stopChan   chan<- struct{}
	reportChan chan<- []string
}

func (a reporter) spawn() *Reporter {
	handle := &Reporter{
		stopChan:   a.stopChan,
		reportChan: a.reportChan,
	}
	go a.loop()
	return handle
}

func (a *reporter) loop() {
	defer func() {
		if err := recover(); err != nil {
			// do something with the error - send to super?
			// a.failure = err
		}
		a.cleanup()
	}()

	a.handleStart()
	for {
		// select over each channel and timer
		select {
		case <-a.stopChan:
			a.handleStop()
			return // special case for Stop
		case m := <-a.reportChan:
			a.handleReport(m)
		}
	}
}

func (a *reporter) cleanup() {
}

func (a *Reporter) Report(name []string) {
	a.reportChan <- name
}
