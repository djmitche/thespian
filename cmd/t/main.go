package main

import (
	"time"

	"github.com/djmitche/thespian"
)

func main() {
	agg := thespian.NewAggregator()

	go func() {
		for _ = range time.NewTicker(900 * time.Millisecond).C {
			agg.Increment("foo")
			agg.Increment("bar")
		}
	}()

	time.Sleep(10 * time.Second)
}
