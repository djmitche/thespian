package gen

import (
	"go/types"
	"strings"
)

const thespianPackage = "github.com/djmitche/thespian"

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
