package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitPackage(t *testing.T) {
	test := func(input, pkg, ident string) func(t *testing.T) {
		return func(t *testing.T) {
			gotPkg, gotIdent := SplitPackage(input)
			require.Equal(t, pkg, gotPkg)
			require.Equal(t, ident, gotIdent)
		}
	}

	t.Run("unqualified type", test("FooBar", "", "FooBar"))
	t.Run("dotted package type", test("foo.bar.FooBar", "foo.bar", "FooBar"))
	t.Run("slashed package type", test("github.com/foo/bar.FooBar", "github.com/foo/bar", "FooBar"))
}
