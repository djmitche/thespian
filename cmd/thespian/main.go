package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

const thespianPackage = "github.com/djmitche/thespian"

var debug = false

func init() {
	debug = os.Getenv("THESPIAN_DEBUG") != ""
}

func dbg(format string, args ...interface{}) {
	if debug {
		log.Printf(format, args...)
	}
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

func initialCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// isFieldOfNamedType determines whether the field is of the given named type.
func isFieldOfNamedType(field *types.Var, pkgPath string, typeName string) bool {
	named, ok := field.Type().(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj.Pkg().Path() != pkgPath {
		return false
	}
	if obj.Name() != typeName {
		return false
	}

	return true
}

// isSendRecvChan returns true if this field is a `chan` without a direction, returning
// a string representing the type of the channel element
func isSendRecvChan(field *types.Var) (bool, string) {
	ch, ok := field.Type().(*types.Chan)
	if !ok {
		return false, ""
	}
	if ch.Dir() != types.SendRecv {
		return false, ""
	}
	return true, ch.Elem().String()
}

// TryActorDefFromObject tries to create an ActorDef from a definition in a package.
func TryActorDefFromObject(pkg *packages.Package, obj types.Object) *actorDef {
	dbg("Examining type %s.%s", pkg.PkgPath, obj.Name())

	// actor implementations are generated from the private struct type, so
	// the object must be private
	if obj.Exported() {
		dbg("..disallowed because exported")
		return nil
	}

	// ..and it must be a type name
	typeName, ok := obj.(*types.TypeName)
	if !ok {
		dbg("..disallowed because not a type name")
		return nil
	}

	// ..and it must be a named type
	namedValue, ok := typeName.Type().(*types.Named)
	if !ok {
		dbg("..disallowed because not a named type")
		return nil
	}

	// ..and it must name a struct
	underlyingStruct, ok := namedValue.Underlying().(*types.Struct)
	if !ok {
		dbg("..disallowed because not a struct")
		return nil
	}

	// ..and that struct must begin with an embedded ActorBase field
	if underlyingStruct.NumFields() < 1 {
		dbg("..disallowed because struct has no fields")
		return nil
	}
	firstField := underlyingStruct.Field(0)

	// ..which must be of type thespianPackage.ActorBase
	dbg("%t %s", firstField.Embedded(), firstField.Name())
	if !firstField.Embedded() || firstField.Name() != "ActorBase" || !isFieldOfNamedType(firstField, thespianPackage, "ActorBase") {
		dbg("..disallowed because struct does not have an embedded %s.ActorBase", thespianPackage)
		return nil
	}

	// ok, this is an actor definition!
	name := typeName.Name()
	def := &actorDef{
		pkg:         pkg,
		privateName: name,
		publicName:  initialCase(name),
		channels:    []channel{},
		timers:      []timer{},
	}

	for i := 0; i < underlyingStruct.NumFields(); i++ {
		field := underlyingStruct.Field(i)
		name := field.Name()
		if isChan, elementType := isSendRecvChan(field); isChan {
			if strings.HasSuffix(name, "Chan") {
				def.channels = append(def.channels, channel{
					name:        name[:len(name)-4],
					elementType: elementType,
				})
			}
		} else if isFieldOfNamedType(field, thespianPackage, "Timer") {
			if strings.HasSuffix(name, "Timer") {
				def.timers = append(def.timers, timer{
					name: name[:len(name)-5],
				})
			}
		}
	}

	return def
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
		ic := initialCase(ch.name)
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
		out.printf("\t\t\ta.handle%s(m)\n", initialCase(ch.name))
	}

	for _, t := range def.timers {
		out.printf("\t\tcase m := <-*a.%sTimer.C:\n", t.name)
		out.printf("\t\t\ta.handle%s(m)\n", initialCase(t.name))
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

// formatter produces formatted Go source
type formatter struct {
	pkg *packages.Package
	buf bytes.Buffer
}

// newFormatter creates a new formatter for the given package
func newFormatter(pkg *packages.Package) *formatter {
	return &formatter{
		pkg: pkg,
	}
}

// add content to the source file
func (f *formatter) printf(format string, args ...interface{}) {
	fmt.Fprintf(&f.buf, format, args...)
}

// write the source file to the package
func (f *formatter) write() {
	dir := filepath.Dir(f.pkg.GoFiles[0])
	filename := filepath.Join(dir, "thespian_generated.go")

	formatted, err := format.Source(f.buf.Bytes())
	if err != nil {
		for i, l := range strings.Split(f.buf.String(), "\n") {
			fmt.Printf("% 4d: %s\n", i+1, l)
		}
		log.Printf("ERROR: generated un-formattable Go: %s", err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(filename, formatted, 0666)
	if err != nil {
		log.Printf("ERROR: could not write %s: %s", filename, err)
		os.Exit(1)
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("thespian: ")

	pkgs, err := packages.Load(&packages.Config{
		Mode:       packages.NeedFiles | packages.NeedName | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Tests:      false,
		BuildFlags: os.Args[1:],
	}, ".")

	if err != nil {
		log.Fatal(err)
	}

	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	pkg := pkgs[0]

	out := newFormatter(pkg)
	out.printf("// code generaged by thespian; DO NOT EDIT\n\n")
	out.printf("package %s\n\n", pkg.Name)
	out.printf("import \"%s\"\n\n", thespianPackage)

	atLeastOne := false
	for _, obj := range pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		def := TryActorDefFromObject(pkg, obj)
		if def != nil {
			def.Generate(out)
			atLeastOne = true
		}
	}

	if !atLeastOne {
		log.Printf("No private actor types found in %s; nothing generated", pkg.PkgPath)
		os.Exit(1)
	}

	out.write()
}
