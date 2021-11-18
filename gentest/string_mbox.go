package gentest

import "github.com/djmitche/thespian"

//go:generate go run ../cmd/thespian mailbox stringMailbox

type stringMailbox struct {
	kind        thespian.SimpleMailbox
	messageType string
}
