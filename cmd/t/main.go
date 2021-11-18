package main

import (
	"time"

	"github.com/djmitche/thespian"
	"github.com/djmitche/thespian/gentest"
)

func main() {
	rt := thespian.NewRuntime()
	rep := gentest.NewReporter(rt)
	agg := gentest.NewAggregator(rt, rep)

	go func() {
		for _ = range time.NewTicker(900 * time.Millisecond).C {
			agg.Increment("foo")
			agg.Increment("bar")
		}
	}()

	time.Sleep(10 * time.Second)
}
