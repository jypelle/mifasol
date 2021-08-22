package config

import (
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restClientV1"
	"golang.org/x/text/collate"
)

type ClientConfig struct {
	*ClientEditableConfig

	Cert     []byte
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
		if clientEditableConfig.Timeout <= 10 {
			clientEditableConfig.Timeout = 10
		} else if clientEditableConfig.Timeout >= 3600 {
			clientEditableConfig.Timeout = 3600
		}

	}

	return &clientEditableConfig
}

func (c *ClientConfig) GetCert() []byte {
	return c.Cert
}

func (c *ClientConfig) SetCert(cert []byte) error {
	c.Cert = cert
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
