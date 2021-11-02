package gentest

import (
	"log"

	"github.com/djmitche/thespian"
)

// lower-case name is the user-provided, internal struct

type reporter struct {
	thespian.ActorBase

	self *Reporter

	// *Chan are treated as message channels
	reportChan chan []string
}

func NewReporter() *Reporter {
	return reporter{
		ActorBase:  thespian.NewActorBase(),
		reportChan: make(chan []string, 5),
	}.spawn()
}

func (a *reporter) handleReport(lines []string) error {
	for _, l := range lines {
		log.Printf("%s\n", l)
	}
	return nil
}
