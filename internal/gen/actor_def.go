package gen

import "golang.org/x/tools/go/packages"

type actorDef struct {
	// package contining the actor definition
	pkg *packages.Package

	// privateName is the name of the private struct
	privateName string

	// publicName is name of the the public struct
	publicName string

	// message channels in this struct
	channels []channel

	// timers in this struct
	timers []timer
}

type channel struct {
	// name of the channel (without "Chan" suffix)
	name string

	// type of the channel elements
	elementType string
}

type timer struct {
	// name of the timer (without "Timer" suffix)
	name string
}

func (def *actorDef) Generate(out *formatter) {
	out.printf("// --- %s\n\n", def.publicName)

	def.generatePublicStruct(out)
	def.generatePublicMethods(out)

	out.printf("// --- %s\n\n", def.privateName)

	def.generateSpawnMethod(out)
	def.generateLoopMethod(out)
	def.generateCleanupMethod(out)
}

func (def *actorDef) generatePublicStruct(out *formatter) {
	out.printf("// %s is the public handle for %s actors.\n", def.publicName, def.privateName)
	out.printf("type %s struct {\n", def.publicName)
	out.printf("\tstopChan chan<- struct{}\n")

	for _, ch := range def.channels {
		out.printf("\t%sChan chan<- %s\n", ch.name, ch.elementType)
	}

	out.printf("}\n\n")
}

func (def *actorDef) generatePublicMethods(out *formatter) {
	out.printf("// Stop sends a message to stop the actor.\n")
	out.printf("func (a *%s) Stop() {\n", def.publicName)
	out.printf("\ta.stopChan <- struct{}{}\n") // TODO: non-blocking send
	out.printf("}\n\n")

	for _, ch := range def.channels {
		ic := publicIdentifier(ch.name)
		out.printf("// %s sends the %s message to the actor.\n", ic, ic)
		out.printf("func (a *%s) %s(m %s) {\n", def.publicName, ic, ch.elementType)
		out.printf("\ta.%sChan <- m\n", ch.name)
		out.printf("}\n\n")
	}
}

func (def *actorDef) generateSpawnMethod(out *formatter) {
	out.printf("func (a %s) spawn(rt *thespian.Runtime) *%s {\n", def.privateName, def.publicName)
	out.printf("\trt.Register(&a.ActorBase)\n")
	out.printf("\thandle := &%s{\n", def.publicName)
	out.printf("\t\tstopChan: a.StopChan,\n")
	for _, ch := range def.channels {
		out.printf("\t\t%sChan: a.%sChan,\n", ch.name, ch.name)
	}
	out.printf("\t}\n")
	out.printf("\tgo a.loop()\n")
	out.printf("\treturn handle\n")
	out.printf("}\n\n")
}

func (def *actorDef) generateLoopMethod(out *formatter) {
	out.printf("func (a *%s) loop() {\n", def.privateName)
	out.printf("\tdefer func() {\n")
	// TODO:
	// out.printf("\t\tif err := recover(); err != nil {}\n")
	out.printf("\t\ta.cleanup()\n")
	out.printf("\t}()\n\n")
	out.printf("\ta.HandleStart()\n")
	out.printf("\tfor {\n")
	out.printf("\t\tselect {\n")
	out.printf("\t\tcase <-a.HealthChan:\n")
	out.printf("\t\t\t// nothing to do\n")
	out.printf("\t\tcase <-a.StopChan:\n")
	out.printf("\t\t\ta.HandleStop()\n")
	out.printf("\t\t\treturn\n")

	for _, ch := range def.channels {
		out.printf("\t\tcase m := <-a.%sChan:\n", ch.name)
		out.printf("\t\t\ta.handle%s(m)\n", publicIdentifier(ch.name))
	}

	for _, t := range def.timers {
		out.printf("\t\tcase m := <-*a.%sTimer.C:\n", t.name)
		out.printf("\t\t\ta.handle%s(m)\n", publicIdentifier(t.name))
	}

	out.printf("\t\t}\n")
	out.printf("\t}\n")
	out.printf("}\n\n")
}

func (def *actorDef) generateCleanupMethod(out *formatter) {
	out.printf("func (a *%s) cleanup() {\n", def.privateName)
	for _, t := range def.timers {
		out.printf("\ta.%sTimer.Stop()\n", t.name)
	}
	out.printf("\ta.Runtime.ActorStopped(&a.ActorBase)\n")
	out.printf("}\n\n")
}
