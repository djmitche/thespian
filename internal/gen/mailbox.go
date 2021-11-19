package gen

import (
	"fmt"
	"text/template"
)

func (yml *MailboxYml) Generate(pkg, name string, out *formatter) error {
	switch yml.Kind {
	case "simple":
		return yml.GenerateSimpleMailbox(pkg, name, out)
	case "ticker":
		return yml.GenerateTickerMailbox(pkg, name, out)
	default:
		return fmt.Errorf("Unknown mailbox kind %s", yml.Kind)
	}
}

func (yml *MailboxYml) GenerateSimpleMailbox(pkg, name string, out *formatter) error {
	vars := struct {
		Pkg         string
		Mailbox     string
		Rx          string
		Tx          string
		MessageType string
	}{
		Pkg:         pkg,
		Mailbox:     name + "Mailbox",
		Rx:          name + "Rx",
		Tx:          name + "Tx",
		MessageType: yml.MessageType,
	}

	var template = template.Must(template.New("simple_mailbox_gen").Funcs(templateFuncs()).Parse(`
// code generaged by thespian; DO NOT EDIT

package {{.Pkg}}

// {{public .Mailbox}} is a mailbox for messages of type {{.MessageType}}.
type {{public .Mailbox}} struct {
	C chan {{.MessageType}}
}

func New{{public .Mailbox}}() {{public .Mailbox}} {
	return {{public .Mailbox}}{
		C: make(chan {{.MessageType}}, 10), // TODO: channel size??
	}
}

// Tx creates a {{.Tx}} for this mailbox
func (mbox *{{public .Mailbox}}) Tx() {{.Tx}} {
	return {{.Tx}}{
		C: mbox.C,
	}
}

// Rx creates a {{.Rx}} for this mailbox
func (mbox *{{public .Mailbox}}) Rx() {{.Rx}} {
	return {{.Rx}}{
		C: mbox.C,
	}
}

// {{.Tx}} sends to a mailbox for messages of type {{.MessageType}}.
type {{.Tx}} struct {
	C chan<- {{.MessageType}}
}

// {{.Rx}} receives from a mailbox for messages of type {{.MessageType}}.
type {{.Rx}} struct {
	C <-chan {{.MessageType}}
}
`))
	out.executeTemplate(template, vars)
	return nil
}

func (yml *MailboxYml) GenerateTickerMailbox(pkg, name string, out *formatter) error {
	vars := struct {
		Pkg     string
		Mailbox string
		Rx      string
	}{
		Pkg:     pkg,
		Mailbox: name + "Mailbox",
		Rx:      name + "Rx",
	}

	var template = template.Must(template.New("ticker_mailbox_gen").Funcs(templateFuncs()).Parse(`
// code generaged by thespian; DO NOT EDIT

package {{.Pkg}}

import "time"

// {{.Rx}} contains a ticker that the actor implementation can control
type {{.Rx}} struct {
	// Ticker is the ticker this mailbox responds to, or nil if it is disabled
	Ticker *time.Ticker
	// Never is a channel that never carries a message, used when Ticker is nil
	never chan time.Time
}

func New{{.Rx}}() {{.Rx}} {
	return {{.Rx}}{
		Ticker: nil,
		// TODO: just use one of these, globally
		never: make(chan time.Time),
	}
}

// Chan gets a channel for this ticker
func (rx *{{.Rx}}) Chan() <-chan time.Time {
	if rx.Ticker != nil {
		return rx.Ticker.C
	}
	return rx.never
}
`))
	out.executeTemplate(template, vars)
	return nil
}
