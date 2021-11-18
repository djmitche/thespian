// code generaged by thespian; DO NOT EDIT

package gentest

// StringMailbox is a mailbox for messages of type string.
type StringMailbox struct {
	C chan string
}

func NewStringMailbox() StringMailbox {
	return StringMailbox{
		C: make(chan string, 10), // TODO: channel size??
	}
}

// Tx creates a StringTx for this mailbox
func (mbox *StringMailbox) Tx() StringTx {
	return StringTx{
		C: mbox.C,
	}
}

// Rx creates a StringRx for this mailbox
func (mbox *StringMailbox) Rx() StringRx {
	return StringRx{
		C: mbox.C,
	}
}

// StringTx sends to a mailbox for messages of type string.
type StringTx struct {
	C chan<- string
}

// StringRx sends to a mailbox for messages of type string.
type StringRx struct {
	C <-chan string
}
