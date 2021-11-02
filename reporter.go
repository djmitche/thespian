package thespian

import (
	"log"
)

// lower-case name is the user-provided, internal struct

type reporter struct {
	AgentBase

	self *Reporter

	// *Chan are treated as message channels
	reportChan chan []string
}

func NewReporter() *Reporter {
	return reporter{
		AgentBase:  NewAgentBase(),
		reportChan: make(chan []string, 5),
	}.spawn()
}

func (a *reporter) handleReport(lines []string) error {
	for _, l := range lines {
		log.Printf("%s\n", l)
	}
	return nil
}
