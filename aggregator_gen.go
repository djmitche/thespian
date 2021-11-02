package thespian

// upper-case name is the external handle
type Aggregator struct {
	stopChan      chan<- struct{}
	incrementChan chan<- string
}

func (a aggregator) spawn() *Aggregator {
	handle := &Aggregator{
		stopChan:      a.stopChan,
		incrementChan: a.incrementChan,
	}
	go a.loop()
	return handle
}

func (a *aggregator) loop() {
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
		case m := <-a.incrementChan:
			a.handleIncrement(m)
		case m := <-*a.flushTimer.c:
			a.handleFlush(m)
		}
	}
}

func (a *aggregator) cleanup() {
	a.flushTimer.Stop()
}

func (a *Aggregator) Increment(name string) {
	a.incrementChan <- name
}
