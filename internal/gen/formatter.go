package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

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
