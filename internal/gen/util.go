package gen

import (
	"bytes"
	"fmt"
	"go/types"
	"os"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/packages"
)

const thespianPackage = "github.com/djmitche/thespian"

// SplitPackage parses a package string of the form package.type into the
// package and type.
func SplitPackage(input string) (string, string) {
	idx := strings.LastIndexByte(input, byte('.'))
	if idx == -1 {
		return "", input
	} else {
		return input[:idx], input[idx+1:]
	}
}

func publicIdentifier(s string) string {
	return strcase.ToCamel(s)
}

func privateIdentifier(s string) string {
	return strcase.ToLowerCamel(s)
}

// Replace the suffix in a string.  If the old suffix does not exist in
// the string, no change occurs and an error is returned.
func swapSuffix(s, old, new string) (string, error) {
	if strings.HasSuffix(s, old) {
		return s[:len(s)-len(old)] + new, nil
	}
	return s, fmt.Errorf("%#v does not have suffix %#v", s, old)
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"public":     publicIdentifier,
		"private":    privateIdentifier,
		"swapSuffix": swapSuffix,
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

func getTemplate(name, tpl string) *template.Template {
	// TODO: caching
	return template.Must(template.New(name).Funcs(templateFuncs()).Parse(tpl))
}

// Render the given template to a string, bailing on any errors.
func renderTemplate(name, tpl string, vars interface{}) string {
	compiled := getTemplate(name, tpl)
	var buf bytes.Buffer
	err := compiled.Execute(&buf, vars)
	if err != nil {
		bail("Error rendering template %s: %s", name, err)
	}
	return buf.String()
}
