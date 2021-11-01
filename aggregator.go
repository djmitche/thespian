package thespian

import (
	"log"
	"time"
)

// lower-case name is the user-provided, internal struct

type aggregator struct {
	AgentBase

	// *Chan are treated as message channels
	incrementChan chan string

	// *Timer are treated as timers
	flushTimer Timer

	counts map[string]int
}

func NewAggregator() *Aggregator {
	return aggregator{
		AgentBase:     NewAgentBase(),
		incrementChan: make(chan string, 5),
		counts:        make(map[string]int),
	}.spawn()
}

func (a *aggregator) handleStart() error {
	log.Printf("start")
	a.flushTimer.Tick(2 * time.Second)
	return nil
}

func (a *aggregator) handleIncrement(name string) error {
	log.Printf("inc %s", name)
	if v, ok := a.counts[name]; ok {
		a.counts[name] = v + 1
	} else {
		a.counts[name] = 1
	}
	return nil
}

func (a *aggregator) handleFlush(t time.Time) error {
	for k, v := range a.counts {
		log.Printf("flush: %s=%d", k, v)
	}
	a.counts = make(map[string]int)
	return nil
}
