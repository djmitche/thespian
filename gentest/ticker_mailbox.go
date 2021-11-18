package gentest

import "github.com/djmitche/thespian"

//go:generate go run ../cmd/thespian mailbox tickerMailbox

type tickerMailbox struct {
	kind        thespian.TickerMailbox
	messageType struct{}
}
