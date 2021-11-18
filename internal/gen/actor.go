// Package gen generates Go code for thespian
package gen

import (
	"go/types"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
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
		publicName:  publicIdentifier(name),
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

func GenerateActor(typeName string) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedFiles | packages.NeedName | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,
	}, ".")

	if err != nil {
		log.Fatal(err)
	}

	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	pkg := pkgs[0]

	typeName = privateIdentifier(typeName)
	var obj types.Object
	for i, o := range pkg.TypesInfo.Defs {
		if i.Name == typeName {
			obj = o
			break
		}
	}
	if obj == nil {
		bail("Type %s not found in this package", typeName)
	}
	def := TryActorDefFromObject(pkg, obj)
	if def == nil {
		bail("Type %s not valid in this package", typeName)
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
