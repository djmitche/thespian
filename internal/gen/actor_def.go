package gen

import (
	"text/template"

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
