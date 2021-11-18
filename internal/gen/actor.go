package gen

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/packages"
)

// TryActorDefFromObject tries to create an ActorDef from a definition in a package.
func TryActorDefFromObject(pkg *packages.Package, obj types.Object) (*actorDef, error) {
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

	return def, nil
}

func GenerateActor(typeName string) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedFiles | packages.NeedName | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,
	}, ".")

	if err != nil {
		bail("Could not parse package: %s", err)
	}

	if len(pkgs) != 1 {
		bail("Parsing package found %d packages (expected 1)", len(pkgs))
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
	def, err := TryActorDefFromObject(pkg, obj)
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
