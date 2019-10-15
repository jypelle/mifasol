package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
)

const configFilename = "config.json"
const configDbFilename = "mifasol.db"
const configDbDirName = "db"
const configDataDirName = "data"
const configSongsDirName = "songs"
const configAlbumsDirName = "albums"
const configAuthorsDirName = "authors"

const configKeyFilename = "key.pem"
const configCertFilename = "cert.pem"

const DefaultPort = 6620
const DefaultSsl = true
const DefaultTimeout = 600

type ServerConfig struct {
	ConfigDir string
	DebugMode bool

	*ServerEditableConfig
}

type ServerEditableConfig struct {
	Hostnames []string `json:"hostnames"`
	Port      int64    `json:"port"`
	Ssl       bool     `json:"ssl"`
	Timeout   int64    `json:"timeout"`
}

func (sc ServerConfig) GetCompleteConfigFilename() string {
	return filepath.Join(sc.ConfigDir, configFilename)
}

func (sc ServerConfig) GetCompleteConfigDbFilename() string {
	return filepath.Join(sc.ConfigDir, configDbFilename)
}

func (sc ServerConfig) GetCompleteConfigDbDirName() string {
	return filepath.Join(sc.ConfigDir, configDbDirName)
}

func (sc ServerConfig) GetCompleteConfigSongsDirName() string {
	return filepath.Join(sc.ConfigDir, configDataDirName, configSongsDirName)
}

func (sc ServerConfig) GetCompleteConfigAlbumsDirName() string {
	return filepath.Join(sc.ConfigDir, configDataDirName, configAlbumsDirName)
}

func (sc ServerConfig) GetCompleteConfigAuthorsDirName() string {
	return filepath.Join(sc.ConfigDir, configDataDirName, configAuthorsDirName)
}

func (sc ServerConfig) GetCompleteConfigKeyFilename() string {
	return filepath.Join(sc.ConfigDir, configKeyFilename)
}

func (sc ServerConfig) GetCompleteConfigCertFilename() string {
	return filepath.Join(sc.ConfigDir, configCertFilename)
}

func NewServerEditableConfig(draftServerEditableConfig *ServerEditableConfig) *ServerEditableConfig {
	var serverEditableConfig ServerEditableConfig

	if draftServerEditableConfig == nil {
		serverEditableConfig = ServerEditableConfig{
			Hostnames: []string{"localhost"},
			Port:      DefaultPort,
			Ssl:       DefaultSsl,
			Timeout:   DefaultTimeout,
		}
	} else {
		serverEditableConfig = *draftServerEditableConfig

		// Check config values
		if serverEditableConfig.Timeout <= 10 {
			serverEditableConfig.Timeout = 10
		} else if serverEditableConfig.Timeout > 3600 {
			serverEditableConfig.Timeout = 3600
		}

	}

	return &serverEditableConfig
}

func (sc *ServerConfig) Save() {
	logrus.Debugf("Save config file: %s", sc.GetCompleteConfigFilename())
	rawConfig, err := json.MarshalIndent(sc.ServerEditableConfig, "", "\t")
	if err != nil {
		logrus.Fatalf("Unable to serialize config file: %v\n", err)
	}
	err = ioutil.WriteFile(sc.GetCompleteConfigFilename(), rawConfig, 0660)
	if err != nil {
		logrus.Fatalf("Unable to save config file: %v\n", err)
	}
}
