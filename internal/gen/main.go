// Package gen generates Go code for thespian
package gen

import (
	"go/types"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

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

func Generate() {
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
