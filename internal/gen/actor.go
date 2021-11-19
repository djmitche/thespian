package gen

import (
	"text/template"
)

func (yml *ActorYml) Generate(pkg, name string, out *formatter) {
	vars := struct {
		Pkg         string
		ThespianPkg string
		Name        string
		Mailboxes   map[string]ActorMailboxYml
	}{
		Pkg:         pkg,
		ThespianPkg: thespianPackage,
		Name:        name,
		Mailboxes:   yml.Mailboxes,
	}
	var template = template.Must(template.New("actor_gen").Funcs(templateFuncs()).Parse(`
// code generaged by thespian; DO NOT EDIT

package {{.Pkg}}

import "{{.ThespianPkg}}"

// --- {{public .Name}}

// {{public .Name}} is the public handle for {{private .Name}} actors.
type {{public .Name}} struct {
	stopChan chan<- struct{}
	{{- range $name, $mbox := .Mailboxes }}
	// TODO: generate this based on the mbox kind
	{{- if eq $mbox.Kind "simple" }}
	{{$name}}Tx {{$mbox.Type}}Tx
	{{- else if eq $mbox.Kind "ticker" }}
	{{- end }}
	{{- end }}
}

// Stop sends a message to stop the actor.  This does not wait until
// the actor has stopped.
func (a *{{public .Name}}) Stop() {
	select {
	case a.stopChan <- struct{}{}:
	default:
	}
}

{{- range $name, $mbox := .Mailboxes }}
{{- if eq $mbox.Kind "simple" }}
// {{public $name}} sends to the actor's {{$name}} mailbox.
func (a *{{public $.Name}}) {{public $name}}(m {{$mbox.MessageType}}) {
	// TODO: generate this based on the mbox kind
	a.{{private $name}}Tx.C <- m
}
{{- else if eq $mbox.Kind "ticker" }}
{{- end }}
{{- end }}

// --- {{private .Name}}

func (a {{private .Name}}) spawn(rt *thespian.Runtime) *{{public .Name}} {
	rt.Register(&a.ActorBase)
	// TODO: these should be in a builder of some sort
	{{- range $name, $mbox := .Mailboxes }}
	{{- if eq $mbox.Kind "simple" }}
	{{$name}}Mailbox := New{{public $mbox.Type}}Mailbox()
	{{- else if eq $mbox.Kind "ticker" }}
	{{- end }}
	{{- end }}

	{{- range $name, $mbox := .Mailboxes }}
	{{- if eq $mbox.Kind "simple" }}
	a.{{$name}}Rx = {{$name}}Mailbox.Rx()
	{{- else if eq $mbox.Kind "ticker" }}
	a.{{$name}}Rx = New{{$mbox.Type}}Rx()
	{{- end }}
	{{- end }}

	handle := &{{public .Name}}{
		stopChan: a.StopChan,
		{{- range $name, $mbox := .Mailboxes }}
		{{- if eq $mbox.Kind "simple" }}
		{{$name}}Tx: {{$name}}Mailbox.Tx(),
		{{- else if eq $mbox.Kind "ticker" }}
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

			{{- range $name, $mbox := .Mailboxes }}
			// TODO: generate this based on the mbox kind
			{{- if eq $mbox.Kind "simple" }}
			case m := <-a.{{$name}}Rx.C:
				a.handle{{public $name}}(m)
			{{- else if eq $mbox.Kind "ticker" }}
			case t := <-a.{{$name}}Rx.Chan():
				a.handle{{public $name}}(t)
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
	out.executeTemplate(template, vars)
}
