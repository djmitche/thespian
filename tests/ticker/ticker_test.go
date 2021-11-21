package ticker

import (
	"fmt"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/djmitche/thespian"
	"github.com/stretchr/testify/require"
)

type timer struct {
	rt *thespian.Runtime
	tx *TimerTx
	rx *TimerRx

	events chan string
	ticks  int
}

func (a *timer) handleStart() {
	a.rx.tkr.Reset(10 * time.Second)
}

func (a *timer) handleStop() {
}

func (a *timer) handleSuperEvent(ev thespian.SuperEvent) {
}

func (a *timer) handleTkr(t time.Time) {
	a.ticks++
	a.events <- fmt.Sprintf("tick @ %s", t)
	if a.ticks == 1 {
		a.rx.tkr.Reset(20 * time.Second)
	} else if a.ticks == 3 {
		a.rx.tkr.Stop()
	}
}

func TestTimer(t *testing.T) {
	clock := clock.NewMock()
	rt := thespian.NewRuntime()
	rt.Clock = clock

	ch := make(chan string, 10)
	tkr := TimerBuilder{timer{events: ch}}.spawn(rt)

	// run the clock for 100s and stop the actor
	go func() {
		for i := 0; i < 100; i++ {
			clock.Add(1 * time.Second)
		}
		tkr.Stop()
	}()

	rt.Run()

	require.Equal(t, "tick @ 1970-01-01 00:00:10 +0000 UTC", <-ch)
	require.Equal(t, "tick @ 1970-01-01 00:00:30 +0000 UTC", <-ch)
	require.Equal(t, "tick @ 1970-01-01 00:00:50 +0000 UTC", <-ch)
	require.Equal(t, 0, len(ch))
}
