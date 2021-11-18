package gentest

//go:generate go run ../cmd/thespian actor aggregator

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

	incrRx  StringRx
	flushRx TickerRx

	// instance vars
	counts     map[string]int
	reporterer Reporterer
}

func NewAggregator(rt *thespian.Runtime, reporterer Reporterer) *Aggregator {
	return aggregator{
		counts:     make(map[string]int),
		reporterer: reporterer,
	}.spawn(rt)
}

func (a *aggregator) HandleStart() error {
	log.Printf("start")
	a.flushRx.Ticker = time.NewTicker(2 * time.Second)
	return nil
}

func (a *aggregator) handleIncr(name string) error {
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
