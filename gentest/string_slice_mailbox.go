package gentest

import "github.com/djmitche/thespian"

//go:generate go run ../cmd/thespian mailbox stringSliceMailbox

type stringSliceMailbox struct {
	kind        thespian.SimpleMailbox
	messageType []string
}
