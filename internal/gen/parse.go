package gen

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

func ParsePackage() (*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedFiles | packages.NeedName | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,
	}, ".")
	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("Parsing package found %d packages (expected 1)", len(pkgs))
	}
	pkg := pkgs[0]

	return pkg, nil
}
