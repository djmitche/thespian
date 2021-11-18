package gentest

//go:generate go run ../cmd/thespian actor reporter

import (
	"log"

	"github.com/djmitche/thespian"
)

// reporter reports aggregated stuff
type reporter struct {
	thespian.ActorBase

	self *Reporter

	// *Chan are treated as message channels
	reportChan chan []string
}

func NewReporter(rt *thespian.Runtime) *Reporter {
	return reporter{
		reportChan: make(chan []string, 5),
	}.spawn(rt)
}

func (a *reporter) handleReport(lines []string) error {
	for _, l := range lines {
		log.Printf("%s\n", l)
	}
	return nil
}
