package gen

import (
	"fmt"
	"go/types"
	"os"
	"text/template"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/packages"
)

const thespianPackage = "github.com/djmitche/thespian"

func publicIdentifier(s string) string {
	return strcase.ToCamel(s)
}

func privateIdentifier(s string) string {
	return strcase.ToLowerCamel(s)
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"public":  publicIdentifier,
		"private": privateIdentifier,
	}
}

// fieldTypeName returns the packge and name of the type of the given field
func fieldTypeName(field *types.Var) (string, string, error) {
	named, ok := field.Type().(*types.Named)
	if !ok {
		return "", "", fmt.Errorf("%s does not have a named type", field)
	}

	obj := named.Obj()
	return obj.Pkg().Path(), obj.Name(), nil
}

// isFieldOfNamedType determines whether the field is of the given named type.
func isFieldOfNamedType(field *types.Var, pkgPath string, typeName string) bool {
	fieldPkgPath, fieldTypeName, err := fieldTypeName(field)
	if err != nil {
		return false
	}
	return fieldPkgPath == pkgPath && fieldTypeName == typeName
}

func findTypeDef(pkg *packages.Package, name string) (types.Object, error) {
	for i, o := range pkg.TypesInfo.Defs {
		if i.Name == name {
			return o, nil
		}
	}
	return nil, fmt.Errorf("Type %s not found in this package", name)
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

// bail prints the message to stderr and exit
func bail(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
