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

// Sender creates a StringSliceSender for this mailbox
func (mbox *StringSliceMailbox) Sender() StringSliceSender {
	return StringSliceSender{
		C: mbox.C,
	}
}

// Receiver creates a StringSliceReceiver for this mailbox
func (mbox *StringSliceMailbox) Receiver() StringSliceReceiver {
	return StringSliceReceiver{
		C: mbox.C,
	}
}

// StringSliceSender sends to a mailbox for messages of type []string.
type StringSliceSender struct {
	C chan<- []string
}

// StringSliceReceiver sends to a mailbox for messages of type []string.
type StringSliceReceiver struct {
	C <-chan []string
}
