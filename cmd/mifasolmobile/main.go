package main

import (
	"flag"
	"fmt"
	"github.com/jypelle/mifasol/internal/mobilecli"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

const configSuffix = "mifasolmobile"

func main() {

	// Logger
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	mainCommand := filepath.Base(os.Args[0])

	// region Flags and Commands definition

	// Debug Mode
	debugMode := flag.Bool("d", false, "Enable debug mode")

	// User config dir
	defaultConfigDir := "./." + configSuffix
	/*
		userConfigDir, err := app.DataDir()
		if err == nil {
			defaultConfigDir = filepath.Join(userConfigDir, configSuffix)
		}
	*/
	configDir := flag.String("c", defaultConfigDir, "Location of mifasolmobile config folder")

	// Usage
	flag.Usage = func() {
		fmt.Printf("\nUsage: %s [OPTIONS] [COMMAND]\n", mainCommand)
		fmt.Printf("\nA self-sufficient music gui for mifasol server\n")
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
	}
	// endregion

	// region Parsing
	flag.Parse()

	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(0)
	}
	// endregion

	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339Nano})
		logrus.Printf("Debug mode activated")
	}

	// Create mifasol mobile client
	mobileApp := mobilecli.NewMobileApp(*configDir, *debugMode)

	// Run mifasol mobile client
	mobileApp.Run()

}
