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
	rt *thespian.Runtime
	tx *AggregatorTx
	rx *AggregatorRx

	// instance vars
	counts     map[string]int
	reporterer Reporterer
}

func NewAggregator(rt *thespian.Runtime, reporterer Reporterer) *AggregatorTx {
	return AggregatorBuilder{
		aggregator: aggregator{
			counts:     make(map[string]int),
			reporterer: reporterer,
		},
	}.spawn(rt)
}

func (a *aggregator) handleStart() {
	log.Printf("start")
	a.rx.flush.Ticker = time.NewTicker(2 * time.Second)
}

func (a *aggregator) handleStop() {
}

func (a *aggregator) handleIncr(name string) {
	log.Printf("inc %s", name)
	if v, ok := a.counts[name]; ok {
		a.counts[name] = v + 1
	} else {
		a.counts[name] = 1
	}
}

func (a *aggregator) handleFlush(t time.Time) {
	lines := []string{}
	for k, v := range a.counts {
		lines = append(lines, fmt.Sprintf("%s=%d", k, v))
	}
	a.reporterer.Report(lines)
	a.counts = make(map[string]int)
}
