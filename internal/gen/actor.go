package gen

import (
	"fmt"
	"go/types"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/packages"
)

type actorDef struct {
	// package contining the actor definition
	Pkg *packages.Package

	// Name is the name of the struct
	Name string

	// MailboxFields enumerates the mailbox fields in this struct
	MailboxFields []MailboxField
}

type MailboxField struct {
	// Name is the name of the field, without the "Sender" or "Receiver" suffix
	Name string

	// Def is the mailboxDef
	Def *MailboxDef
}

// NewActorDef creates a new actor definition based on a type with the given
// private name.
func NewActorDef(pkg *packages.Package, name string) (*actorDef, error) {
	obj, err := findTypeDef(pkg, name)
	if err != nil {
		return nil, err
	}

	// actor implementations are generated from the private struct type, so
	// the object must be private
	if obj.Exported() {
		return nil, fmt.Errorf("type is exported")
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

	// ..and that struct must begin with an embedded ActorBase field
	if underlyingStruct.NumFields() < 1 {
		return nil, fmt.Errorf("struct has no fields")
	}
	firstField := underlyingStruct.Field(0)

	// ..which must be of type thespianPackage.ActorBase
	if !firstField.Embedded() || firstField.Name() != "ActorBase" || !isFieldOfNamedType(firstField, thespianPackage, "ActorBase") {
		return nil, fmt.Errorf("struct does not have an embedded %s.ActorBase", thespianPackage)
	}

	// ok, this is an actor definition!
	def := &actorDef{
		Pkg:  pkg,
		Name: name,
	}

	for i := 0; i < underlyingStruct.NumFields(); i++ {
		field := underlyingStruct.Field(i)
		name := field.Name()
		if strings.HasSuffix(name, "Receiver") {
			_, typeName, err := fieldTypeName(field)
			if err != nil {
				continue
			}
			mboxDef, err := NewMailboxDefForReceiver(pkg, typeName)
			if err != nil {
				return nil, fmt.Errorf("Invalid type for %s: %s", name, err)
			}
			def.MailboxFields = append(def.MailboxFields, MailboxField{
				Name: name[:len(name)-len("Receiver")],
				Def:  mboxDef,
			})
		}
	}

	return def, nil
}

func (def *actorDef) Generate(out *formatter) {
	var template = template.Must(template.New("actor_gen").Funcs(templateFuncs()).Parse(`
// code generaged by thespian; DO NOT EDIT

package {{.Pkg.Name}}

// TODO: use variable
import "github.com/djmitche/thespian"

// --- {{public .Name}}

// {{public .Name}} is the public handle for {{private .Name}} actors.
type {{public .Name}} struct {
	stopChan chan<- struct{}
	{{- range .MailboxFields }}
	// TODO: generate this based on the mbox kind
	{{- if eq .Def.Kind "SimpleMailbox" }}
	{{private .Name}}Sender {{swapSuffix .Def.Name "Mailbox" "Sender" | public}}
	{{- else if eq .Def.Kind "TickerMailbox" }}
	{{- end }}
	{{- end }}
}

// Stop sends a message to stop the actor.
func (a *{{public .Name}}) Stop() {
	a.stopChan <- struct{}{}
}

{{ range .MailboxFields }}
{{- if eq .Def.Kind "SimpleMailbox" }}
// {{public .Name}} sends to the actor's {{public .Name}} mailbox.
func (a *{{public $.Name}}) {{public .Name}}(m {{.Def.MessageType}}) {
	// TODO: generate this based on the mbox kind
	a.{{private .Name}}Sender.C <- m
}
{{- else if eq .Def.Kind "TickerMailbox" }}
{{- end }}
{{- end }}

// --- {{private .Name}}

func (a {{private .Name}}) spawn(rt *thespian.Runtime) *{{public .Name}} {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	{{- range .MailboxFields }}
	// TODO: generate based on mbox kind
	{{- if eq .Def.Kind "SimpleMailbox" }}
	{{private .Name}}Mailbox := New{{public .Def.Name}}()
	{{- else if eq .Def.Kind "TickerMailbox" }}
	{{- end }}
	{{- end }}

	{{- range .MailboxFields }}
	// TODO: generate based on mbox kind
	{{- if eq .Def.Kind "SimpleMailbox" }}
	a.{{private .Name}}Receiver = {{private .Name}}Mailbox.Receiver()
	{{- else if eq .Def.Kind "TickerMailbox" }}
	a.{{private .Name}}Receiver = New{{swapSuffix .Def.Name "Mailbox" "Receiver" | public }}()
	{{- end }}
	{{- end }}

	handle := &{{public .Name}}{
		stopChan: a.StopChan,
		{{- range .MailboxFields }}
		// TODO: generate based on mbox kind
		{{- if eq .Def.Kind "SimpleMailbox" }}
		{{private .Name}}Sender: {{private .Name}}Mailbox.Sender(),
		{{- else if eq .Def.Kind "TickerMailbox" }}
		{{- end }}
		{{- end }}
	}
	go a.loop()
	return handle
}

func (a *{{private .Name}}) loop() {
	defer func() {
		a.cleanup()
	}()
	a.HandleStart()
	for {
		select {
		case <-a.HealthChan:
			// nothing to do
		case <-a.StopChan:
			a.HandleStop()
			return

			{{- range .MailboxFields }}
			// TODO: generate this based on the mbox kind
			{{- if eq .Def.Kind "SimpleMailbox" }}
			case m := <-a.{{private .Name}}Receiver.C:
				a.handle{{public .Name}}(m)
			{{- else if eq .Def.Kind "TickerMailbox" }}
			case t := <-a.{{private .Name}}Receiver.Chan():
				a.handle{{public .Name}}(t)
			{{- end }}
			{{- end }}
		}
	}
}

func (a *{{private .Name}}) cleanup() {
	// TODO: clean up mboxes too
	a.Runtime.ActorStopped(&a.ActorBase)
}
`))
	out.executeTemplate(template, def)
}

func GenerateActor(typeName string) {
	pkg, err := ParsePackage()
	if err != nil {
		bail("Could not parse package: %s", err)
	}

	def, err := NewActorDef(pkg, typeName)
	if err != nil {
		bail("Could not build actor for type %s: %s", typeName, err)
	}

	out := newFormatter(pkg, strcase.ToSnake(typeName)+"_thespian_gen.go")
	def.Generate(out)
	err = out.write()
	if err != nil {
		bail("Error: %s", err)
	}
}
