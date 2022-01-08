package config

import (
	"encoding/json"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/collate"
	"io/ioutil"
	"os"
	"path/filepath"
)

const configFilename = "config.json"
const configCertFilename = "cert.pem"

type ClientConfig struct {
	ConfigDir string
	DebugMode bool

	*ClientEditableConfig

	collator *collate.Collator
}

func (c *ClientConfig) Collator() *collate.Collator {
	if c.collator == nil {
		c.collator = collate.New(tool.LocaleTags[c.SortLanguage])
	}
	return c.collator
}

type ClientEditableConfig struct {
	ServerHostname   string `json:"serverHostname"`
	ServerPort       int64  `json:"serverPort"`
	ServerSsl        bool   `json:"serverSsl"`
	ServerSelfSigned bool   `json:"serverSelfSigned"`
	SortLanguage     string `json:"sortLanguage"`
	BufferLength     int64  `json:"bufferLength"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Timeout          int64  `json:"timeout"`
}

func NewClientEditableConfig(draftClientEditableConfig *ClientEditableConfig) *ClientEditableConfig {
	var clientEditableConfig ClientEditableConfig

	if draftClientEditableConfig == nil {
		clientEditableConfig = ClientEditableConfig{
			ServerHostname:   restClientV1.DefaultServerHostname,
			ServerPort:       restClientV1.DefaultServerPort,
			ServerSsl:        restClientV1.DefaultServerSsl,
			ServerSelfSigned: restClientV1.DefaultServerSelfSigned,
			SortLanguage:     restClientV1.DefaultSortLanguage,
			BufferLength:     restClientV1.DefaultBufferLength,
			Username:         restClientV1.DefaultUsername,
			Password:         restClientV1.DefaultPassword,
			Timeout:          restClientV1.DefaultTimeout,
		}
	} else {
		clientEditableConfig = *draftClientEditableConfig

		// Check config values
		if _, ok := tool.LocaleTags[clientEditableConfig.SortLanguage]; !ok {
			clientEditableConfig.SortLanguage = restClientV1.DefaultSortLanguage
		}
		if clientEditableConfig.BufferLength <= 10 {
			clientEditableConfig.BufferLength = 10
		} else if clientEditableConfig.BufferLength > 5000 {
			clientEditableConfig.BufferLength = 5000
		}
		if clientEditableConfig.Timeout <= 10 {
			clientEditableConfig.Timeout = 10
		} else if clientEditableConfig.Timeout >= 3600 {
			clientEditableConfig.Timeout = 3600
		}

	}

	return &clientEditableConfig
}

func (c *ClientConfig) Save() {
	logrus.Debugf("Save config file: %s", c.GetCompleteConfigFilename())
	rawConfig, err := json.MarshalIndent(c.ClientEditableConfig, "", "\t")
	if err != nil {
		logrus.Fatalf("Unable to serialize config file: %v\n", err)
	}
	err = ioutil.WriteFile(c.GetCompleteConfigFilename(), rawConfig, 0660)
	if err != nil {
		logrus.Fatalf("Unable to save config file: %v\n", err)
	}
}

func (c *ClientConfig) GetCompleteConfigFilename() string {
	return filepath.Join(c.ConfigDir, configFilename)
}

func (c *ClientConfig) GetCert() []byte {
	certPem, err := os.ReadFile(filepath.Join(c.ConfigDir, configCertFilename))
	if err != nil {
		return nil
	}

	return certPem
}

func (c *ClientConfig) SetCert(cert []byte) error {
	err := os.WriteFile(filepath.Join(c.ConfigDir, configCertFilename), cert, 0660)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientConfig) GetServerHostname() string {
	return c.ServerHostname
}

func (c *ClientConfig) GetServerPort() int64 {
	return c.ServerPort
}

func (c *ClientConfig) GetServerSsl() bool {
	return c.ServerSsl
}

func (c *ClientConfig) GetServerSelfSigned() bool {
	return c.ServerSelfSigned
}

func (c *ClientConfig) GetTimeout() int64 {
	return c.Timeout
}

func (c *ClientConfig) GetUsername() string {
	return c.Username
}

func (c *ClientConfig) GetPassword() string {
	return c.Password
}
