// code generaged by thespian; DO NOT EDIT

package gentest

// StringSliceMailbox is a mailbox for messages of type []string.
type StringSliceMailbox struct {
	C chan []string
}

func NewStringSliceMailbox() StringSliceMailbox {
	return StringSliceMailbox{
		C: make(chan []string, 10), // TODO: channel size??
	}
}

// Tx creates a StringSliceTx for this mailbox
func (mbox *StringSliceMailbox) Tx() StringSliceTx {
	return StringSliceTx{
		C: mbox.C,
	}
}

// Rx creates a StringSliceRx for this mailbox
func (mbox *StringSliceMailbox) Rx() StringSliceRx {
	return StringSliceRx{
		C: mbox.C,
	}
}

// StringSliceTx sends to a mailbox for messages of type []string.
type StringSliceTx struct {
	C chan<- []string
}

// StringSliceRx sends to a mailbox for messages of type []string.
type StringSliceRx struct {
	C <-chan []string
}
