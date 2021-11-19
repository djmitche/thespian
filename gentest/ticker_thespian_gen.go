// code generaged by thespian; DO NOT EDIT

package gentest

import "time"

// TickerRx contains a ticker that the actor implementation can control
type TickerRx struct {
	// Ticker is the ticker this mailbox responds to, or nil if it is disabled
	Ticker *time.Ticker
	// Never is a channel that never carries a message, used when Ticker is nil
	never chan time.Time
}

func NewTickerRx() TickerRx {
	return TickerRx{
		Ticker: nil,
		// TODO: just use one of these, globally
		never: make(chan time.Time),
	}
}

// Chan gets a channel for this ticker
func (rx *TickerRx) Chan() <-chan time.Time {
	if rx.Ticker != nil {
		return rx.Ticker.C
	}
	return rx.never
}
