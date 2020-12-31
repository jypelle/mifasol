package srv

import (
	"fmt"
	"os"
)

func (c *ServerApp) Config(
	hostnames []string,
	port int64,
	ssl *bool) {

	shouldSaveConfig := false

	if len(hostnames) > 0 {
		c.ServerEditableConfig.Hostnames = hostnames
		os.Remove(c.GetCompleteConfigKeyFilename())
		os.Remove(c.GetCompleteConfigCertFilename())
		shouldSaveConfig = true
		fmt.Println("Server hostnames updated")
		fmt.Println("Self-signed certificate will be regenerated, clients should accept the new one")
	}

	if port > 0 {
		c.ServerEditableConfig.Port = port
		shouldSaveConfig = true
		fmt.Println("Server port updated")
	}

	if ssl != nil {
		c.ServerEditableConfig.Ssl = *ssl
		shouldSaveConfig = true
		if *ssl {
			fmt.Println("SSL enabled: clients should use https to connect to server")
		} else {
			fmt.Println("SSL disabled: clients should use http to connect to server")
		}
	}

	if shouldSaveConfig {
		c.ServerConfig.Save()
	}
}
