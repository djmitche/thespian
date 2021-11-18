package gen

import (
	"log"
	"os"
)

var debug = false

func init() {
	debug = os.Getenv("THESPIAN_DEBUG") != ""
}

func dbg(format string, args ...interface{}) {
	if debug {
		log.Printf(format, args...)
	}
}
