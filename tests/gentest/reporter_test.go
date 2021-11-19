package gentest

import (
	"log"

	"github.com/djmitche/thespian"
)

// reporter reports aggregated stuff
type reporter struct {
	rt *thespian.Runtime
	tx *ReporterTx
	rx *ReporterRx
}

func NewReporter(rt *thespian.Runtime) *ReporterTx {
	return ReporterBuilder{}.spawn(rt)
}

func (a *reporter) handleStart() {
}

func (a *reporter) handleStop() {
}

func (a *reporter) handleReport(lines []string) {
	for _, l := range lines {
		log.Printf("%s\n", l)
	}
}
