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

	// message Channels in this struct
	Channels []channel

	// Timers in this struct
	Timers []timer
}

type channel struct {
	// PublicName is the public name of the channel (without "Chan" suffix)
	PublicName string

	// PrivateName is the private name of the channel (without "Chan" suffix)
	PrivateName string

	// type of the channel elements
	ElementType string
}

type timer struct {
	// PublicName is the public name of the timer (without "Timer" suffix)
	PublicName string

	// PrivateName is the private name of the timer (without "Timer" suffix)
	PrivateName string
}

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
		Channels:    []channel{},
		Timers:      []timer{},
	}

	for i := 0; i < underlyingStruct.NumFields(); i++ {
		field := underlyingStruct.Field(i)
		name := field.Name()
		if isChan, elementType := isSendRecvChan(field); isChan {
			if strings.HasSuffix(name, "Chan") {
				def.Channels = append(def.Channels, channel{
					PublicName:  publicIdentifier(name[:len(name)-4]),
					PrivateName: privateIdentifier(name[:len(name)-4]),
					ElementType: elementType,
				})
			}
		} else if isFieldOfNamedType(field, thespianPackage, "Timer") {
			if strings.HasSuffix(name, "Timer") {
				def.Timers = append(def.Timers, timer{
					PublicName:  publicIdentifier(name[:len(name)-5]),
					PrivateName: privateIdentifier(name[:len(name)-5]),
				})
			}
		}
	}

	return def, nil
}

func (def *actorDef) Generate(out *formatter) {
	var template = template.Must(template.New("actor_gen").Parse(`
// --- {{.PublicName}}

// {{.PublicName}} is the public handle for {{.PrivateName}} actors.
type {{.PublicName}} struct {
	stopChan chan<- struct{}
	{{- range .Channels }}
	{{.PrivateName}}Chan chan<- {{.ElementType}}
	{{- end }}
}

// Stop sends a message to stop the actor.
func (a *{{.PublicName}}) Stop() {
	a.stopChan <- struct{}{}
}

{{ range .Channels }}
// {{.PublicName}} sends the {{.PublicName}} message to the actor.
func (a *{{$.PublicName}}) {{.PublicName}}(m {{.ElementType}}) {
	a.{{.PrivateName}}Chan <- m
}
{{- end }}

// --- {{.PrivateName}}

func (a {{.PrivateName}}) spawn(rt *thespian.Runtime) *{{.PublicName}} {
	rt.Register(&a.ActorBase)
	handle := &{{.PublicName}}{
		stopChan: a.StopChan,
		{{- range .Channels }}
			{{.PrivateName}}Chan: a.{{.PrivateName}}Chan,
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

			{{- range .Channels }}
			case m := <-a.{{.PrivateName}}Chan:
				a.handle{{.PublicName}}(m)
			{{- end }}

			{{- range .Timers }}
			case m := <-*a.{{.PrivateName}}Timer.C:
				a.handle{{.PublicName}}(m)
			{{- end }}
		}
	}
}

func (a *{{.PrivateName}}) cleanup() {
	{{- range .Timers }}
	a.{{.PrivateName}}Timer.Stop()
	{{- end }}
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
