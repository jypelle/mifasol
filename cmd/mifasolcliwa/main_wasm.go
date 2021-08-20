package main

import (
	"github.com/jypelle/mifasol/internal/cliwa"
)

func main() {

	// Create mifasol webassembly client
	app := cliwa.NewApp(true)

	// Run
	app.Start()

}
