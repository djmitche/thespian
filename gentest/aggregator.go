package gentest

import (
	"fmt"
	"log"
	"time"

	"github.com/djmitche/thespian"
)

// (naming things is hard)
type Reporterer interface {
	Report([]string)
}

// lower-case name is the user-provided, internal struct
type aggregator struct {
	thespian.ActorBase

	// self reference
	self *Aggregator

	// *Chan are treated as message channels
	incrementChan chan string

	// *Timer are treated as timers
	flushTimer thespian.Timer

	// instance vars
	counts     map[string]int
	reporterer Reporterer
}

func NewAggregator(reporterer Reporterer) *Aggregator {
	return aggregator{
		ActorBase:     thespian.NewActorBase(),
		incrementChan: make(chan string, 5),
		counts:        make(map[string]int),
		reporterer:    reporterer,
	}.spawn()
}

func (a *aggregator) HandleStart() error {
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
	lines := []string{}
	for k, v := range a.counts {
		lines = append(lines, fmt.Sprintf("%s=%d", k, v))
	}
	a.reporterer.Report(lines)
	a.counts = make(map[string]int)
	return nil
}
