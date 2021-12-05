package main

import (
	"flag"
	"fmt"
	"github.com/jypelle/mifasol/internal/srv"
	"github.com/jypelle/mifasol/internal/version"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const configSuffix = "mifasolsrv"

func main() {
	// Set random seed
	rand.Seed(time.Now().UnixNano())

	// Logger
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	mainCommand := filepath.Base(os.Args[0])

	// region flags and Commands definition

	// Debug Mode
	debugMode := flag.Bool("d", false, "Enable debug mode")

	// User config dir
	defaultConfigDir := "./." + configSuffix
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		defaultConfigDir = filepath.Join(userConfigDir, configSuffix)
	}
	configDir := flag.String("c", defaultConfigDir, "Location of mifasolsrv config folder")

	// Usage
	flag.Usage = func() {
		fmt.Printf("\nUsage: %s [OPTIONS] [COMMAND]\n", mainCommand)
		fmt.Printf("\nA self-hosted music server\n")
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
		fmt.Printf("\nCommands:\n")
		fmt.Printf("  config    Configure server\n")
		fmt.Printf("  run       Run server\n")
		fmt.Printf("  version   Show the version number\n")
		fmt.Printf("\nRun '%s COMMAND --help' for more information on a command.\n", mainCommand)
	}

	// config command
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	configHostnames := configCmd.String("hostnames", "", "Set comma separated hostname list used to generate self-signed certificate")
	configPort := configCmd.Int64("n", 0, "Set port number")
	//configRenewSelfSignedCertificate := configCmd.Bool("renew-sscrt", false, "Renew self-signed certificate")
	configSslEnabled := configCmd.Bool("enable-ssl", false, "Enable SSL with self-signed certificate (client should use https to connect to server)")
	configSslDisabled := configCmd.Bool("disable-ssl", false, "Disable SSL (client should use http to connect to server)")

	configCmd.Usage = func() {
		fmt.Printf("\nUsage: %s config\n", mainCommand)
		fmt.Printf("\nConfigure the mifasol server\n")
		fmt.Printf("\nOptions:\n")
		configCmd.PrintDefaults()
	}

	// run command
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)

	runCmd.Usage = func() {
		fmt.Printf("\nUsage: %s run\n", mainCommand)
		fmt.Printf("\nRun the server\n")
	}

	// version command
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	versionCmd.Usage = func() {
		fmt.Printf("\nUsage: %s version\n", mainCommand)
		fmt.Printf("\nShow the mifasol server version information\n")
	}

	// endregion

	// region Parsing
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	switch flag.Arg(0) {
	case "config":
		configCmd.Parse(flag.Args()[1:])
		if configCmd.NArg() > 0 {
			fmt.Printf("\n\"%s %s\" accepts no arguments\n", mainCommand, flag.Arg(0))
			configCmd.Usage()
			os.Exit(1)
		}
		if configCmd.NFlag() == 0 {
			fmt.Printf("\n\"%s %s\" should provide at least one flag\n", mainCommand, flag.Arg(0))
			configCmd.Usage()
			os.Exit(1)
		}
	case "run":
		runCmd.Parse(flag.Args()[1:])
		if runCmd.NArg() > 0 {
			fmt.Printf("\n\"%s %s\" accepts no arguments\n", mainCommand, flag.Arg(0))
			runCmd.Usage()
			os.Exit(1)
		}
	case "version":
		versionCmd.Parse(flag.Args()[1:])
		if versionCmd.NArg() > 0 {
			fmt.Printf("\n\"%s %s\" accepts no arguments\n", mainCommand, flag.Arg(0))
			versionCmd.Usage()
			os.Exit(1)
		}
	default:
		fmt.Printf("\n%s is not a mifasolsrv command\n", flag.Args()[0])
		flag.Usage()
		os.Exit(1)
	}
	// endregion

	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339Nano})
		logrus.Printf("Debug mode activated")
	}

	// Create mifasol server
	serverApp := srv.NewServerApp(*configDir, *debugMode)

	if configCmd.Parsed() {
		// Update mifasol server config
		var configSsl *bool = nil
		if *configSslEnabled {
			trueVar := true
			configSsl = &trueVar
		}
		if *configSslDisabled {
			falseVar := false
			configSsl = &falseVar
		}

		hostnames := strings.Split(strings.ReplaceAll(*configHostnames, " ", ""), ",")

		serverApp.Config(
			hostnames,
			*configPort,
			configSsl)

	} else if versionCmd.Parsed() {
		fmt.Printf("Version %s\n", version.AppVersion.String())
	} else {
		if runCmd.Parsed() {
			// Start mifasol server
			serverApp.Start()
			defer serverApp.Stop()

			// Listen stop signal
			ch := make(chan os.Signal)
			signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGHUP)
			sig := <-ch
			logrus.Printf("Received signal: %v", sig)
		}
	}

}
