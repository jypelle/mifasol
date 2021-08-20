package cli

import (
	"fmt"
)

func (c *ClientApp) Config(
	serverHostname string,
	serverPort int64,
	serverSsl *bool,
	serverSelfSignedCertificate *bool,
	username string,
	password string,
	clearCachedServerCertificate bool) {
	shouldSaveConfig := false

	if serverHostname != "" {
		c.config.ClientEditableConfig.ServerHostname = serverHostname
		shouldSaveConfig = true
		fmt.Println("Server host updated")
	}

	if serverPort > 0 {
		c.config.ClientEditableConfig.ServerPort = serverPort
		shouldSaveConfig = true
		fmt.Println("Server port updated")
	}

	if serverSsl != nil {
		c.config.ClientEditableConfig.ServerSsl = *serverSsl
		shouldSaveConfig = true
		if *serverSsl {
			fmt.Println("SSL enabled: mifasolcli will use https to connect to server")
		} else {
			fmt.Println("SSL disabled: mifasolcli will use http to connect to server")
		}
	}

	if serverSelfSignedCertificate != nil {
		c.config.ClientEditableConfig.ServerSelfSigned = *serverSelfSignedCertificate
		shouldSaveConfig = true
		if *serverSelfSignedCertificate {
			fmt.Println("Self-signed server authorized")
		} else {
			fmt.Println("Self-signed server not authorized")
		}
	}

	if username != "" {
		c.config.ClientEditableConfig.Username = username
		shouldSaveConfig = true
		fmt.Println("Username updated")
	}

	if password != "" {
		c.config.ClientEditableConfig.Password = password
		shouldSaveConfig = true
		fmt.Println("Password updated")
	}

	if clearCachedServerCertificate {
		c.config.SetCert(nil)
		fmt.Println("Cached server certificate has been deleted")
	}

	if shouldSaveConfig {
		c.config.Save()
	}
}
