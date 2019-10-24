package cli

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/jypelle/mifasol/cli/config"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/jypelle/mifasol/tool"
	"github.com/jypelle/mifasol/version"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type ClientApp struct {
	config     config.ClientConfig
	restClient *restClientV1.RestClient
}

func NewClientApp(configDir string, debugMode bool) *ClientApp {

	logrus.Debugf("Creation of mifasol client %s ...", version.AppVersion.String())

	app := &ClientApp{
		config: config.ClientConfig{
			ConfigDir: configDir,
			DebugMode: debugMode,
		},
	}

	// Check Configuration folder
	_, err := os.Stat(app.config.ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Printf("Creation of config folder: %s", app.config.ConfigDir)
			err = os.Mkdir(app.config.ConfigDir, 0770)
			if err != nil {
				logrus.Fatalf("Unable to create config folder: %v\n", err)
			}
		} else {
			logrus.Fatalf("Unable to access config folder: %s", app.config.ConfigDir)
		}
	}

	// Open configuration file
	var draftClientEditableConfig *config.ClientEditableConfig

	rawConfig, err := ioutil.ReadFile(app.config.GetCompleteConfigFilename())
	if err == nil {
		// Interpret configuration file
		draftClientEditableConfig = &config.ClientEditableConfig{}
		err = json.Unmarshal(rawConfig, draftClientEditableConfig)
		if err != nil {
			logrus.Fatalf("Unable to interpret config file: %v\n", err)
		}
	}

	app.config.ClientEditableConfig = config.NewClientEditableConfig(draftClientEditableConfig)

	app.config.Save()

	logrus.Debugln("Client created")

	return app
}

func (c *ClientApp) Init() {
	// Check server certificate
	if c.config.ServerSsl && c.config.ServerSelfSigned {
		existServerCert, err := tool.IsFileExists(c.config.GetCompleteConfigCertFilename())
		if err != nil {
			logrus.Fatalf("Unable to access %s: %v\n", c.config.GetCompleteConfigCertFilename(), err)
		}
		if !existServerCert {
			// Retrieve & store server certificate
			insecureTr := &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}

			insecureClient := &http.Client{
				Transport: insecureTr,
				Timeout:   time.Second * time.Duration(c.config.Timeout),
			}

			// Prepare the request
			req, err := http.NewRequest("GET", c.getServerUrl()+"/isalive", nil)
			if err != nil {
				logrus.Fatalf("Unable to connect to mifasol server: %v\n", err)
			}

			// Send the request
			response, err := insecureClient.Do(req)
			if err != nil {
				logrus.Fatalf("Unable to connect to mifasol server: %v\n", err)
			}
			defer response.Body.Close()

			if len(response.TLS.PeerCertificates) == 0 {
				logrus.Fatalf("Unable to connect to mifasol server: certificate is missing\n", err)
			}

			// Retrieve server certificate
			cert := response.TLS.PeerCertificates[0]

			// Save server certificate
			tool.CertToFile(c.config.GetCompleteConfigCertFilename(), cert.Raw)
		}

		// Check secure connection
		certPem, err := ioutil.ReadFile(c.config.GetCompleteConfigCertFilename())
		if err != nil {
			logrus.Fatalf("Reading server certificate failed : %v", err)
		}

		rootCAPool := x509.NewCertPool()
		rootCAPool.AppendCertsFromPEM(certPem)

		secureTr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            rootCAPool,
			},
		}

		secureClient := &http.Client{
			Transport: secureTr,
			Timeout:   time.Second * time.Duration(c.config.Timeout),
		}

		// Prepare the request
		req, err := http.NewRequest("GET", c.getServerUrl()+"/isalive", nil)
		if err != nil {
			logrus.Fatalf("Unable to prepare mifasol server connection: %v\n", err)
		}

		// Send the request
		response, err := secureClient.Do(req)
		if err != nil {
			if urlErr, ok := err.(*url.Error); ok {
				if hostnameError, ok := urlErr.Err.(x509.HostnameError); ok {
					logrus.Fatalf("Bad hostname: Mifasol server is available but should be reconfigured to accept connection with \"%s\" hostname\n", hostnameError.Host)
				}
				if _, ok := urlErr.Err.(x509.UnknownAuthorityError); ok {
					fmt.Print("Mifasol server certificate has changed.\nDo you accept the new one (to prevent man-in-the-middle attack, you should explicitely accept) ?[y/N]\n")
					reader := bufio.NewReader(os.Stdin)
					text, _ := reader.ReadString('\n')
					text = strings.Replace(text, "\n", "", -1)
					text = strings.Replace(text, "\r", "", -1)
					if text == "y" || text == "Y" {
						os.Remove(c.config.GetCompleteConfigCertFilename())
						c.Init()
						return
					} else {
						fmt.Print("New server certificate has been refused.")
						os.Exit(1)
					}
				}
				if _, ok := urlErr.Err.(x509.CertificateInvalidError); ok {
					logrus.Fatalf("Invalid certificate: Mifasol server is available but should regenerate its SSL certificate.\n")
				}
			}
			logrus.Fatalf("Unable to connect to mifasol server: %v\n", err)
		} else {
			defer response.Body.Close()
		}

	}

	// Create rest Client
	var e error
	c.restClient, e = restClientV1.NewRestClient(&c.config)
	if e != nil {
		logrus.Fatalf("Unable to instanciate mifasol rest client: %v\n", e)
	}
}

func (c *ClientApp) getServerUrl() string {
	if c.config.ServerSsl {
		return "https://" + c.config.ServerHostname + ":" + strconv.FormatInt(c.config.ServerPort, 10)
	} else {
		return "http://" + c.config.ServerHostname + ":" + strconv.FormatInt(c.config.ServerPort, 10)
	}
}
