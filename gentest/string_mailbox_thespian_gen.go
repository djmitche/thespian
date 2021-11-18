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

// Sender creates a StringSender for this mailbox
func (mbox *StringMailbox) Sender() StringSender {
	return StringSender{
		C: mbox.C,
	}
}

// Receiver creates a StringReceiver for this mailbox
func (mbox *StringMailbox) Receiver() StringReceiver {
	return StringReceiver{
		C: mbox.C,
	}
}

// StringSender sends to a mailbox for messages of type string.
type StringSender struct {
	C chan<- string
}

// StringReceiver sends to a mailbox for messages of type string.
type StringReceiver struct {
	C <-chan string
}
