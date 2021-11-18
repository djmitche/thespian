package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// formatter produces formatted Go source
type formatter struct {
	filename string
	buf      bytes.Buffer
}

// newFormatter creates a new formatter for the given package
func newFormatter(pkg *packages.Package, filename string) *formatter {
	dir := filepath.Dir(pkg.GoFiles[0])
	filename = filepath.Join(dir, filename)

	return &formatter{
		filename: filename,
	}
}

// add content to the source file
func (f *formatter) printf(format string, args ...interface{}) {
	fmt.Fprintf(&f.buf, format, args...)
}

// write the source file to the package
func (f *formatter) write() error {
	formatted, err := format.Source(f.buf.Bytes())
	if err != nil {
		for i, l := range strings.Split(f.buf.String(), "\n") {
			fmt.Printf("% 4d: %s\n", i+1, l)
		}
		return fmt.Errorf("generated un-formattable Go: %s", err)
	}
	err = ioutil.WriteFile(f.filename, formatted, 0666)
	if err != nil {
		return fmt.Errorf("could not write %s: %s", f.filename, err)
	}
	return nil
}
