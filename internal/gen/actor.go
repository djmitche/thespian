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
	pkg *packages.Package

	// PrivateName is the name of the private struct
	PrivateName string

	// PublicName is name of the the public struct
	PublicName string

	// Timers in this struct
	Timers []timer

	// Mailboxes in this struct
	Mailboxes []mailbox
}

type timer struct {
	// PublicName is the public name of the timer (without "Timer" suffix)
	PublicName string

	// PrivateName is the private name of the timer (without "Timer" suffix)
	PrivateName string
}

type mailbox struct {
	// Name is the name of the field, without the "Sender" or "Receiver" suffix
	Name string

	// Def is the mailboxDef
	Def *mailboxDef
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
		pkg:         pkg,
		PrivateName: name,
		PublicName:  publicIdentifier(name),
		Timers:      []timer{},
	}

	for i := 0; i < underlyingStruct.NumFields(); i++ {
		field := underlyingStruct.Field(i)
		name := field.Name()
		if isFieldOfNamedType(field, thespianPackage, "Timer") {
			if strings.HasSuffix(name, "Timer") {
				def.Timers = append(def.Timers, timer{
					PublicName:  publicIdentifier(name[:len(name)-5]),
					PrivateName: privateIdentifier(name[:len(name)-5]),
				})
			}
		}

		if strings.HasSuffix(name, "Receiver") {
			_, typeName, err := fieldTypeName(field)
			if err != nil {
				continue
			}
			mboxDef, err := NewMailboxDefForReceiver(pkg, typeName)
			if err != nil {
				return nil, fmt.Errorf("Invalid type for %s: %s", name, err)
			}
			def.Mailboxes = append(def.Mailboxes, mailbox{
				Name: name[:len(name)-len("Receiver")],
				Def:  mboxDef,
			})
		}
	}

	return def, nil
}

func (def *actorDef) Generate(out *formatter) {
	var template = template.Must(template.New("actor_gen").Funcs(templateFuncs()).Parse(`
// --- {{.PublicName}}

// {{.PublicName}} is the public handle for {{.PrivateName}} actors.
type {{.PublicName}} struct {
	stopChan chan<- struct{}
	{{- range .Mailboxes }}
	{{private .Name}}Sender {{swapSuffix .Def.Name "Mailbox" "Sender" | public}}
	{{- end }}
}

// Stop sends a message to stop the actor.
func (a *{{.PublicName}}) Stop() {
	a.stopChan <- struct{}{}
}

{{ range .Mailboxes }}
// {{public .Name}} sends to the actor's {{public .Name}} mailbox.
func (a *{{$.PublicName}}) {{public .Name}}(m {{.Def.MessageType}}) {
	// TODO: generate this based on the mbox kind
	a.{{private .Name}}Sender.C <- m
}
{{- end }}

// --- {{.PrivateName}}

func (a {{.PrivateName}}) spawn(rt *thespian.Runtime) *{{.PublicName}} {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	{{- range .Mailboxes }}
	// TODO: generate based on mbox kind
	{{private .Name}}Mailbox := New{{public .Def.Name}}()
	{{- end }}

	{{- range .Mailboxes }}
	// TODO: generate based on mbox kind
	a.{{private .Name}}Receiver = {{private .Name}}Mailbox.Receiver()
	{{- end }}

	handle := &{{.PublicName}}{
		stopChan: a.StopChan,
		{{- range .Mailboxes }}
			// TODO: generate based on mbox kind
			{{private .Name}}Sender: {{private .Name}}Mailbox.Sender(),
		{{- end }}
	}
	go a.loop()
	return handle
}

func (a *{{.PrivateName}}) loop() {
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

			{{- range .Timers }}
			case m := <-*a.{{.PrivateName}}Timer.C:
				a.handle{{.PublicName}}(m)
			{{- end }}

			{{- range .Mailboxes }}
			// TODO: generate this based on the mbox kind
			case m := <-a.{{private .Name}}Receiver.C:
				a.handle{{public .Name}}(m)
			{{- end }}
		}
	}
}

func (a *{{.PrivateName}}) cleanup() {
	{{- range .Timers }}
	a.{{.PrivateName}}Timer.Stop()
	{{- end }}
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
	out.printf("// code generaged by thespian; DO NOT EDIT\n\n")
	out.printf("package %s\n\n", pkg.Name)
	out.printf("import \"%s\"\n\n", thespianPackage)
	def.Generate(out)
	err = out.write()
	if err != nil {
		bail("Error: %s", err)
	}
}
