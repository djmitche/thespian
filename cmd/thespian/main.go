package main

import (
	"log"

	"github.com/djmitche/thespian/internal/gen"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("thespian: ")

	gen.Generate()
}
