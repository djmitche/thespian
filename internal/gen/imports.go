package gen

import (
	"fmt"
	"strings"
)

// importTracker tracks importTracker in a module, generating unique names for each one.
type importTracker struct {
	// formatted imports
	imports []string

	// map from fully-qualified name to local name
	names map[string]string
}

func newImportTracker() *importTracker {
	return &importTracker{
		names: make(map[string]string),
	}
}

// Add an import, generating and returning a local name for it
func (imp *importTracker) add(pkg string) string {
	if shortName, found := imp.names[pkg]; found {
		return shortName
	}
	shortName := fmt.Sprintf("import%d", len(imp.imports))
	imp.imports = append(imp.imports, fmt.Sprintf(`%s "%s"`, shortName, pkg))
	imp.names[pkg] = shortName

	return shortName
}

// Add an import with the given (hard-coded) short name.
func (imp *importTracker) addNamed(pkg, shortName string) {
	if existingShortName, found := imp.names[pkg]; found {
		if existingShortName != shortName {
			panic(fmt.Sprintf("package %s already defined as %s", pkg, existingShortName))
		}
		return
	}
	if strings.HasSuffix(pkg, "/"+shortName) {
		imp.imports = append(imp.imports, fmt.Sprintf(`"%s"`, pkg))
	} else {
		imp.imports = append(imp.imports, fmt.Sprintf(`%s "%s"`, shortName, pkg))
	}
	imp.names[pkg] = shortName
}

// Get the list of imports for the Go file
func (imp *importTracker) get() []string {
	return imp.imports
}
