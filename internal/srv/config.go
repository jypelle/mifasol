package srv

import (
	"fmt"
	"os"
)

func (s *ServerApp) Config(
	hostnames []string,
	port int64,
	ssl *bool) {

	shouldSaveConfig := false

	if len(hostnames) > 0 {
		s.ServerEditableConfig.Hostnames = hostnames
		os.Remove(s.GetCompleteConfigKeyFilename())
		os.Remove(s.GetCompleteConfigCertFilename())
		shouldSaveConfig = true
		fmt.Println("Server hostnames updated")
		fmt.Println("Self-signed certificate will be regenerated, clients should accept the new one")
	}

	if port > 0 {
		s.ServerEditableConfig.Port = port
		shouldSaveConfig = true
		fmt.Println("Server port updated")
	}

	if ssl != nil {
		s.ServerEditableConfig.Ssl = *ssl
		shouldSaveConfig = true
		if *ssl {
			fmt.Println("SSL enabled: clients should use https to connect to server")
		} else {
			fmt.Println("SSL disabled: clients should use http to connect to server")
		}
	}

	if shouldSaveConfig {
		s.ServerConfig.Save()
	}
}
