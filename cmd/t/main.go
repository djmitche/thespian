package main

import (
	"time"

	"github.com/djmitche/thespian/gentest"
)

func main() {
	rep := gentest.NewReporter()
	agg := gentest.NewAggregator(rep)

	go func() {
		for _ = range time.NewTicker(900 * time.Millisecond).C {
			agg.Increment("foo")
			agg.Increment("bar")
		}
	}()

	time.Sleep(10 * time.Second)
}
