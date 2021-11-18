package thespian

// SimpleMailbox simply wraps a fixed-size Go channel.  Messages go in, messags
// come out.
type SimpleMailbox struct{}

// TickerMailbox wraps a Ticker, sending empty messages at a fixed interval.
type TickerMailbox struct{}
