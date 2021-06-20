package main

import (
	"log"

	"github.com/shaunschembri/restreamer/internal/restreamer"
)

var version string

func main() {
	log.Printf("Starting restreamer %s", version)
	restreamer.Main()
}
