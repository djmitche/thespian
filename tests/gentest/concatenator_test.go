package gentest

import (
	"log"
	"strings"

	"github.com/djmitche/thespian"
)

// concatenator reports aggregated stuff
type concatenator struct {
	rt *thespian.Runtime
	tx *ConcatenatorTx
	rx *ConcatenatorRx

	accumulator []string
}

func NewConcatenator(rt *thespian.Runtime) *ConcatenatorTx {
	return ConcatenatorBuilder{}.spawn(rt)
}

func (a *concatenator) handleStart() {
}

func (a *concatenator) handleStop() {
}

func (a *concatenator) handleInput(v string) {
	log.Printf("input: %s", v)
	a.accumulator = append(a.accumulator, v)
}

func (a *concatenator) handleOutput(ch chan<- string) {
	log.Printf("output")
	ch <- strings.Join(a.accumulator, "")
	a.accumulator = nil
}
