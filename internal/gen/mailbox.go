package gen

import (
	"fmt"
	"go/types"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/packages"
)

type MailboxKind = string

const (
	SimpleMailboxKind MailboxKind = "SimpleMailbox"
	TickerMailboxKind MailboxKind = "TickerMailbox"
)

// Parse a string as a mailbox kind
func toMailboxKind(name string) (MailboxKind, error) {
	switch name {
	case SimpleMailboxKind:
		return MailboxKind(name), nil
	case TickerMailboxKind:
		return MailboxKind(name), nil
	default:
		return "", fmt.Errorf("Unknown mailbox kind %s", name)
	}
}

type mailboxDef struct {
	// package contining the mailbox definition
	Pkg *packages.Package

	// Name is the name of the mailbox struct
	Name string

	// Kind gives the kind of this mailbox
	Kind MailboxKind

	// MessageType is the type of messages in this mailbox
	MessageType string
}

// NewMailboxDefForReceiver creates a new mailboxDef for the given receiver type name
func NewMailboxDefForReceiver(pkg *packages.Package, name string) (*mailboxDef, error) {
	mailboxName := privateIdentifier(strings.Replace(name, "Receiver", "Mailbox", 1))
	return NewMailboxDef(pkg, mailboxName)
}

// NewMailboxDef creates a new mailbox definition based on a type with the given
// private name.
func NewMailboxDef(pkg *packages.Package, name string) (*mailboxDef, error) {
	obj, err := findTypeDef(pkg, name)
	if err != nil {
		return nil, err
	}

	// mailbox implementations are generated from the private struct type, so
	// the object must be private
	if obj.Exported() {
		return nil, fmt.Errorf("type is exported")
	}

	// ..and must end in Mailbo
	if !strings.HasSuffix(name, "Mailbox") {
		return nil, fmt.Errorf("type name does not end in 'Mailbox'")
	}

	// ..and it must be a type name
	typeName, ok := obj.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("object is not a type name")
	}

	// ..and it must be a named type
	namedValue, ok := typeName.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("object is not a named type")
	}

	// ..and it must name a struct
	underlyingStruct, ok := namedValue.Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("object is not a struct")
	}

	// ..and that struct have exactly the right shape
	if underlyingStruct.NumFields() != 2 {
		return nil, fmt.Errorf("struct must have two fields: kind and messageType")
	}

	// first field must be named "kind"
	firstField := underlyingStruct.Field(0)
	if firstField.Name() != "kind" {
		return nil, fmt.Errorf("struct's first field is not named 'kind'")
	}

	// ..and must be of type thespianPackage.<kind>
	firstFieldPkg, firstFieldTypeName, err := fieldTypeName(firstField)
	if err != nil {
		return nil, fmt.Errorf("cannot examine struct's first field: %s", err)
	}
	if firstFieldPkg != thespianPackage {
		return nil, fmt.Errorf("struct's first field is not from %s", thespianPackage)
	}

	// ..and must be a valid Mailbox kind
	mailboxKind, err := toMailboxKind(firstFieldTypeName)
	if err != nil {
		return nil, fmt.Errorf("Unrecognized embedded struct field: %s", err)
	}

	// the second field must be a message type
	secondField := underlyingStruct.Field(1)
	if secondField.Name() != "messageType" {
		return nil, fmt.Errorf("struct's second field is not named 'messageType'")
	}

	messageType := secondField.Type().String()

	// ok, this is an mailbox definition!
	def := &mailboxDef{
		Pkg:         pkg,
		Name:        name,
		Kind:        mailboxKind,
		MessageType: messageType,
	}

	return def, nil
}

func (def *mailboxDef) Generate(out *formatter) {
	switch def.Kind {
	case SimpleMailboxKind:
		def.GenerateSimpleMailbox(out)
	case TickerMailboxKind:
		def.GenerateTickerMailbox(out)
	default:
		panic("undefined kind")
	}
}

func (def *mailboxDef) GenerateSimpleMailbox(out *formatter) {
	var template = template.Must(template.New("simple_mailbox_gen").Funcs(templateFuncs()).Parse(`
{{- $sender := swapSuffix .Name "Mailbox" "Sender" | public }}
{{- $receiver := swapSuffix .Name "Mailbox" "Receiver" | public }}
// code generaged by thespian; DO NOT EDIT

package {{.Pkg.Name}}

// {{public .Name}} is a mailbox for messages of type {{.MessageType}}.
type {{public .Name}} struct {
	C chan {{.MessageType}}
}

func New{{public .Name}}() {{public .Name}} {
	return {{public .Name}}{
		C: make(chan {{.MessageType}}, 10), // TODO: channel size??
	}
}

// Sender creates a {{$sender}} for this mailbox
func (mbox *{{public .Name}}) Sender() {{$sender}} {
	return {{$sender}}{
		C: mbox.C,
	}
}

// Receiver creates a {{$receiver}} for this mailbox
func (mbox *{{public .Name}}) Receiver() {{$receiver}} {
	return {{$receiver}}{
		C: mbox.C,
	}
}

// {{$sender}} sends to a mailbox for messages of type {{.MessageType}}.
type {{$sender}} struct {
	C chan<- {{.MessageType}}
}

// {{$receiver}} sends to a mailbox for messages of type {{.MessageType}}.
type {{$receiver}} struct {
	C <-chan {{.MessageType}}
}
`))
	out.executeTemplate(template, def)
}

func (def *mailboxDef) GenerateTickerMailbox(out *formatter) {
	var template = template.Must(template.New("ticker_mailbox_gen").Funcs(templateFuncs()).Parse(`
{{- $receiver := swapSuffix .Name "Mailbox" "Receiver" | public }}
// code generaged by thespian; DO NOT EDIT

package {{.Pkg.Name}}

import "time"

// {{$receiver}} sends to a mailbox for messages of type {{.MessageType}}.
type {{$receiver}} struct {
	// Ticker is the ticker this mailbox responds to, or nil if it is disabled
	Ticker *time.Ticker
	// Never is a channel that never carries a message
	never chan time.Time
}

{{- $receiver := swapSuffix .Name "Mailbox" "Receiver" | public }}
func New{{$receiver}}() {{$receiver}} {
	return {{$receiver}}{
		Ticker: nil,
		// TODO: just use one of these, globally
		never: make(chan time.Time),
	}
}

// Chan gets a channel for this ticker
func (rx *{{$receiver}}) Chan() <-chan time.Time {
	if rx.Ticker != nil {
		return rx.Ticker.C
	}
	return rx.never
}
`))
	out.executeTemplate(template, def)
}

func GenerateMailbox(typeName string) {
	pkg, err := ParsePackage()
	if err != nil {
		bail("Could not parse package: %s", err)
	}

	def, err := NewMailboxDef(pkg, typeName)
	if err != nil {
		bail("Could not build mailbox for type %s: %s", typeName, err)
	}

	out := newFormatter(pkg, strcase.ToSnake(typeName)+"_thespian_gen.go")
	def.Generate(out)
	err = out.write()
	if err != nil {
		bail("Error: %s", err)
	}
}
