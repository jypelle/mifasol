package main

import (
	"github.com/jypelle/mifasol/internal/cliwa"
	"math/rand"
	"time"
)

func main() {
	// Set random seed
	rand.Seed(time.Now().UnixNano())

	// Create mifasol webassembly client
	app := cliwa.NewApp(true)

	// Run
	app.Start()

}
