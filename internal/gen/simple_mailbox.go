package gen

import "strings"

type SimpleMailboxGenerator struct {
	ThisPackage  string
	MessageType  string
	MboxTypeBase string
	Imports      []string

	// actor generation only
	MboxTypeQual string
	ActorName    string
	FieldName    string
}

func (g *SimpleMailboxGenerator) GenerateGo(out *formatter) {
	out.executeTemplate(getTemplate(
		"simple_mailbox_go_file", `
// code generated by thespian; DO NOT EDIT

package {{.ThisPackage}}

import (
	{{- range .Imports }}
	{{ . }}
	{{- end }}
)

// {{public .MboxTypeBase}}Mailbox is a mailbox for messages of type {{.MessageType}}.
type {{public .MboxTypeBase}}Mailbox struct {
	// C is the bidirectional channel over which messages will be transferred.  If
	// this is not set in the mailbox, a fresh channel will be created automatically.
	C chan {{.MessageType}}
	// Disabled, if set to true, causes the mailbox to start life disabled.
	Disabled bool
}

// ApplyDefaults applies default settings to this {{public .MboxTypeBase}}, if
// the struct has its zero value.
func (mbox *{{public .MboxTypeBase}}Mailbox) ApplyDefaults() {
	if mbox.C == nil {
		mbox.C = make(chan {{.MessageType}}, 10) // default channel size
	}
}

// Tx creates a {{.MboxTypeBase}}Tx for this mailbox
func (mbox *{{public .MboxTypeBase}}Mailbox) Tx() {{.MboxTypeBase}}Tx {
	return {{.MboxTypeBase}}Tx{
		C: mbox.C,
	}
}

// Rx creates a {{.MboxTypeBase}}Rx for this mailbox
func (mbox *{{public .MboxTypeBase}}Mailbox) Rx() {{.MboxTypeBase}}Rx {
	return {{.MboxTypeBase}}Rx{
		C: mbox.C,
		Disabled: mbox.Disabled,
	}
}

// {{.MboxTypeBase}}Tx sends to a mailbox for messages of type {{.MessageType}}.
type {{.MboxTypeBase}}Tx struct {
	C chan<- {{.MessageType}}
}

// {{.MboxTypeBase}}Rx receives from a mailbox for messages of type {{.MessageType}}.
type {{.MboxTypeBase}}Rx struct {
	C <-chan {{.MessageType}}
	// Disabled, if set to true, will disable receipt of messages from this mailbox.
	Disabled bool
}

// Chan gets a channel for this mailbox, or nil if there is nothing to select from.
func (rx *{{.MboxTypeBase}}Rx) Chan() <-chan {{.MessageType}} {
	if rx.Disabled {
		return nil
	}
	return rx.C
}`), g)
}

func (g *SimpleMailboxGenerator) ActorBuilderStructDecl() string {
	return renderTemplate(
		"simple_actor_builder_struct_decl",
		`{{.FieldName}} {{.MboxTypeQual}}{{.MboxTypeBase}}Mailbox`,
		g)
}

func (g *SimpleMailboxGenerator) ActorRxStructDecl() string {
	return renderTemplate(
		"simple_actor_rx_struct_decl",
		`{{.FieldName}} {{.MboxTypeQual}}{{.MboxTypeBase}}Rx`,
		g)
}

func (g *SimpleMailboxGenerator) ActorRxInitializer() string {
	return renderTemplate(
		"simple_actor_rx_initializer",
		`{{.FieldName}}: bldr.{{.FieldName}}.Rx(),`,
		g)
}

func (g *SimpleMailboxGenerator) ActorTxStructDecl() string {
	return renderTemplate(
		"simple_actor_tx_struct_decl",
		`{{.FieldName}} {{.MboxTypeQual}}{{.MboxTypeBase}}Tx`,
		g)
}

func (g *SimpleMailboxGenerator) ActorTxInitializer() string {
	return renderTemplate(
		"simple_actor_tx_initializer",
		`{{.FieldName}}: bldr.{{.FieldName}}.Tx(),`,
		g)
}

func (g *SimpleMailboxGenerator) ActorTxStructMethod() string {
	return renderTemplate(
		"simple_actor_tx_struct_method", strings.TrimSpace(`
		// {{public .FieldName}} sends to the actor's {{.FieldName}} mailbox.
		func (tx *{{public .ActorName}}Tx) {{public .FieldName}}(m {{.MessageType}}) {
			tx.{{private .FieldName}}.C <- m
		}`), g)
}

func (g *SimpleMailboxGenerator) ActorSpawnSetupClause() string {
	return renderTemplate(
		"simple_actor_spawn_setup_clause", strings.TrimSpace(`
		bldr.{{.FieldName}}.ApplyDefaults()
		`), g)
}

func (g *SimpleMailboxGenerator) ActorLoopCase() string {
	return renderTemplate(
		"simple_actor_loop_case", strings.TrimSpace(`
			case m := <-rx.{{.FieldName}}.Chan():
				a.handle{{public .FieldName}}(m)
		`), g)
}

func (g *SimpleMailboxGenerator) ActorCleanupClause() string {
	return ""
}
