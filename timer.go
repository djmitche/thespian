package thespian

import "time"

type Timer struct {
	C      *<-chan time.Time
	ticker *time.Ticker
}

func (t *Timer) Tick(dur time.Duration) {
	t.ticker = time.NewTicker(dur)
	t.C = &t.ticker.C
}

func (t *Timer) Stop() {
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
	t.C = nil
}
