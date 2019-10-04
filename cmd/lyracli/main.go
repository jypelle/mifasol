package main

import (
	"flag"
	"fmt"
	"lyra/cli"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

const configSuffix = "lyracli"

func main() {

	// Logger
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	mainCommand := filepath.Base(os.Args[0])

	// region Flags and Commands definition

	// Debug Mode
	debugMode := flag.Bool("d", false, "Enable debug mode")

	// User config dir
	defaultConfigDir := "./." + configSuffix
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		defaultConfigDir = filepath.Join(userConfigDir, configSuffix)
	}
	configDir := flag.String("c", defaultConfigDir, "Location of lyracli config folder")

	// Usage
	flag.Usage = func() {
		fmt.Printf("\nUsage: %s [OPTIONS] [COMMAND]\n", mainCommand)
		fmt.Printf("\nA self-sufficient music client for lyra server\n")
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
		fmt.Printf("\nCommands:\n")
		fmt.Printf("  config    Configure the lyra client\n")
		fmt.Printf("  ui        Launch the console interface\n")
		fmt.Printf("  import    Import every flac, mp3 and ogg files from current folder to lyra server\n")
		fmt.Printf("  filesync  Sync a folder with lyra server content\n")
		fmt.Printf("\nRun '%s COMMAND --help' for more information on a command.\n", mainCommand)
	}

	// config command
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	configServerHostname := configCmd.String("hostname", "", "Set server host name")
	configServerPort := configCmd.Int64("n", 0, "Set server port number")
	configUsername := configCmd.String("u", "", "Set username")
	configPassword := configCmd.String("p", "", "Set password")
	configClearCachedSelfSignedServerCertificate := configCmd.Bool("clear-sscrt", false, "Clear cached self-signed server certificate")
	configServerSSLEnabled := configCmd.Bool("enable-ssl", false, "Enable SSL (use https to connect to server)")
	configServerSSLDisabled := configCmd.Bool("disable-ssl", false, "Disable SSL (use http to connect to server)")
	configServerSelfSignedCertificateAccepted := configCmd.Bool("accept-sscrt", false, "Accept Self-signed server certificate")
	configServerSelfSignedCertificateRefused := configCmd.Bool("refuse-sscrt", false, "Refuse Self-signed server certificate")

	configCmd.Usage = func() {
		fmt.Printf("\nUsage: %s config [OPTIONS]\n", mainCommand)
		fmt.Printf("\nConfigure the lyra client\n")
		fmt.Printf("\nOptions:\n")
		configCmd.PrintDefaults()
	}

	// ui command
	uiCmd := flag.NewFlagSet("ui", flag.ExitOnError)

	uiCmd.Usage = func() {
		fmt.Printf("\nUsage: %s ui\n", mainCommand)
		fmt.Printf("\nLaunch the console interface\n")
	}

	// import command
	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	importOneFolderPerAlbumDisabled := importCmd.Bool("disable-one-folder-per-album", false, "Don't use folder name changes to differentiate homonym albums")

	importCmd.Usage = func() {
		fmt.Printf("\nUsage: %s import [OPTIONS] [Location of music folder to import]\n", mainCommand)
		fmt.Printf("\nImport flac, mp3 and ogg files to lyra server\n")
		fmt.Printf("\nOptions:\n")
		configCmd.PrintDefaults()
	}

	// filesync command
	fileSyncCmd := flag.NewFlagSet("filesync", flag.ExitOnError)

	fileSyncCmd.Usage = func() {
		fmt.Printf("\nUsage: %s filesync [SUBCOMMAND]\n", mainCommand)
		fmt.Printf("\nSync a folder with lyra server content\n")
		fmt.Printf("\nSubcommands:\n")
		fmt.Printf("  init    Prepare folder for synchronization with lyra server\n")
		fmt.Printf("  sync    Synchronize folder with lyra server content\n")
	}

	// filesync init subcommand
	fileSyncCmdInitSubCmd := flag.NewFlagSet("init", flag.ExitOnError)

	fileSyncCmdInitSubCmd.Usage = func() {
		fmt.Printf("\nUsage: %s %s %s [Location of folder to synchronize]\n", mainCommand, fileSyncCmd.Name(), fileSyncCmdInitSubCmd.Name())
		fmt.Printf("\nPrepare music folder for synchronization with lyra server\n")
	}

	// filesync sync subcommand
	fileSyncCmdSyncSubCmd := flag.NewFlagSet("sync", flag.ExitOnError)

	fileSyncCmdSyncSubCmd.Usage = func() {
		fmt.Printf("\nUsage: %s %s %s [Location of folder to synchronize]\n", mainCommand, fileSyncCmd.Name(), fileSyncCmdSyncSubCmd.Name())
		fmt.Printf("\nSynchronize music folder content with lyra server\n")
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
	case "ui":
		uiCmd.Parse(flag.Args()[1:])
		if uiCmd.NArg() > 0 {
			fmt.Printf("\n\"%s %s\" accepts no arguments\n", mainCommand, flag.Arg(0))
			uiCmd.Usage()
			os.Exit(1)
		}
	case "import":
		importCmd.Parse(flag.Args()[1:])
		if importCmd.NArg() != 1 {
			fmt.Printf("\n\"%s %s\" need a folder to import\n", mainCommand, flag.Arg(0))
			importCmd.Usage()
			os.Exit(1)
		}
	case "filesync":
		fileSyncCmd.Parse(flag.Args()[1:])
		if fileSyncCmd.NArg() < 1 {
			fmt.Printf("\n\"%s %s\" need a subcommand\n", mainCommand, flag.Arg(0))
			fileSyncCmd.Usage()
			os.Exit(1)
		}

		switch fileSyncCmd.Arg(0) {
		case "init":
			fileSyncCmdInitSubCmd.Parse(fileSyncCmd.Args()[1:])
			if fileSyncCmdInitSubCmd.NArg() != 1 {
				fmt.Printf("\n\"%s %s %s\" need a folder to initialize\n", mainCommand, flag.Arg(0), fileSyncCmd.Arg(0))
				fileSyncCmdInitSubCmd.Usage()
				os.Exit(1)
			}
		case "sync":
			fileSyncCmdSyncSubCmd.Parse(fileSyncCmd.Args()[1:])
			if fileSyncCmdSyncSubCmd.NArg() != 1 {
				fmt.Printf("\n\"%s %s %s\" need a folder to synchronize\n", mainCommand, flag.Arg(0), fileSyncCmd.Arg(0))
				fileSyncCmdSyncSubCmd.Usage()
				os.Exit(1)
			}
		default:
			fmt.Printf("\n%s is not a lyracli %s subcommand\n", fileSyncCmd.Arg(0), flag.Arg(0))
			fileSyncCmd.Usage()
			os.Exit(1)
		}
	default:
		fmt.Printf("\n%s is not a lyracli command\n", flag.Args()[0])
		flag.Usage()
		os.Exit(1)
	}
	// endregion

	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339Nano})
		logrus.Printf("Debug mode activated")
	}

	// Create lyra client
	clientApp := cli.NewClientApp(*configDir, *debugMode)

	if configCmd.Parsed() {

		// Update lyra client config
		var configServerSSL *bool = nil
		if *configServerSSLEnabled {
			trueVar := true
			configServerSSL = &trueVar
		}
		if *configServerSSLDisabled {
			falseVar := false
			configServerSSL = &falseVar
		}

		var configServerSelfSignedCertificate *bool = nil
		if *configServerSelfSignedCertificateAccepted {
			trueVar := true
			configServerSelfSignedCertificate = &trueVar
		}
		if *configServerSelfSignedCertificateRefused {
			falseVar := false
			configServerSelfSignedCertificate = &falseVar
		}

		clientApp.Config(
			*configServerHostname,
			*configServerPort,
			configServerSSL,
			configServerSelfSignedCertificate,
			*configUsername,
			*configPassword,
			*configClearCachedSelfSignedServerCertificate)

	} else {

		// Init lyra client
		clientApp.Init()

		if uiCmd.Parsed() {
			// Start console user interface
			clientApp.UI()
		}

		if importCmd.Parsed() {
			// Import songs
			clientApp.Import(importCmd.Arg(0), *importOneFolderPerAlbumDisabled)
		}

		if fileSyncCmd.Parsed() {
			if fileSyncCmdInitSubCmd.Parsed() {
				// Prepare music folder for synchronisation
				clientApp.FileSyncInit(fileSyncCmdInitSubCmd.Arg(0))
			}

			if fileSyncCmdSyncSubCmd.Parsed() {
				// Sync music folder
				clientApp.FileSyncSync(fileSyncCmdSyncSubCmd.Arg(0))
			}
		}

	}

}
