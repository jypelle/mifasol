package restClientV1

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/internal/version"
	"github.com/jypelle/mifasol/restApiV1"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const JsonContentType = "application/json"

var (
	ErrBadHostname        = fmt.Errorf("Bad hostname: Mifasol server is available but should be reconfigured to accept connection with specified hostname")
	ErrBadCertificate     = fmt.Errorf("Mifasol server certificate has changed")
	ErrInvalidCertificate = fmt.Errorf("Invalid certificate: Mifasol server is available but should regenerate its SSL certificate.")
)

type RestClient struct {
	ClientConfig       RestConfig
	httpClient         *http.Client
	token              *restApiV1.Token
	webassemblyEnabled bool
}

func NewRestClient(clientConfig RestConfig, webassemblyEnabled bool) (*RestClient, error) {

	var rootCAPool *x509.CertPool = nil

	// Load self-signed server certificate
	if clientConfig.GetServerSsl() && clientConfig.GetServerSelfSigned() && !webassemblyEnabled {

		// Define Root CA
		rootCAPool = x509.NewCertPool()

		// Load local server certificate
		if len(clientConfig.GetCert()) == 0 {
			// First connection to mifasol server: retrieve & store self-signed server certificate
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
				Timeout:   time.Second * time.Duration(clientConfig.GetTimeout()),
			}

			// Prepare the request
			req, err := http.NewRequest("GET", getServerUrl(clientConfig)+"/isalive", nil)
			if err != nil {
				return nil, fmt.Errorf("Unable to connect to mifasol server: %v", err)
			}

			// Send the request
			response, err := insecureClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("Unable to connect to mifasol server: %v", err)
			}
			defer response.Body.Close()

			if response.TLS == nil || len(response.TLS.PeerCertificates) == 0 {
				return nil, fmt.Errorf("Unable to connect to mifasol server: certificate is missing")
			}

			// Retrieve & save server certificate
			err = clientConfig.SetCert(tool.CertToMemory(response.TLS.PeerCertificates[0].Raw))
			if err != nil {
				return nil, fmt.Errorf("Unable to store mifasol server certificate: %v", err)
			}
		}

		if len(clientConfig.GetCert()) == 0 {
			return nil, fmt.Errorf("Reading server certificate failed")
		}

		// Append server certificate to root CAs
		rootCAPool.AppendCertsFromPEM(clientConfig.GetCert())

	}

	// Configure client
	tr := &http.Transport{
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

	restClient := &RestClient{
		ClientConfig: clientConfig,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   time.Second * time.Duration(clientConfig.GetTimeout()),
		},
		webassemblyEnabled: webassemblyEnabled,
	}

	// Check secure connection

	// Prepare the request
	req, err := http.NewRequest("GET", getServerUrl(clientConfig)+"/isalive", nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare mifasol server connection: %v\n", err)
	}

	// Send the request
	response, err := restClient.httpClient.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if _, ok := urlErr.Err.(x509.HostnameError); ok {
				return nil, ErrBadHostname
			}
			if _, ok := urlErr.Err.(x509.UnknownAuthorityError); ok {
				return nil, ErrBadCertificate
			}
			if _, ok := urlErr.Err.(x509.CertificateInvalidError); ok {
				return nil, ErrInvalidCertificate
			}
		}
		return nil, fmt.Errorf("Unable to connect to mifasol server: %v", err)
	} else {
		defer response.Body.Close()
	}

	return restClient, nil
}

func getServerUrl(restConfig RestConfig) string {
	if restConfig.GetServerSsl() {
		return "https://" + restConfig.GetServerHostname() + ":" + strconv.FormatInt(restConfig.GetServerPort(), 10)
	} else {
		return "http://" + restConfig.GetServerHostname() + ":" + strconv.FormatInt(restConfig.GetServerPort(), 10)
	}
}

func (c *RestClient) getServerApiUrl() string {
	return getServerUrl(c.ClientConfig) + "/api/v1"
}

// doRequest prepare and send an http request, managing access token renewal for expired token
func (c *RestClient) doRequest(method, relativeUrl string, contentType string, body io.Reader) (*http.Response, ClientError) {

	// Dear mifasolsrv, could you gimme a token ?
	_, cliErr := c.GetToken()
	if cliErr != nil {
		return nil, cliErr
	}

	// Prepare the request
	var req *http.Request
	var err error
	if method == "GET" && body != nil && c.webassemblyEnabled {
		req, err = http.NewRequest("POST", c.getServerApiUrl()+relativeUrl, body)
		if err != nil {
			return nil, NewClientError(err)
		}
		req.Header.Add("x-http-method-override", "GET")
	} else {
		req, err = http.NewRequest(method, c.getServerApiUrl()+relativeUrl, body)
		if err != nil {
			return nil, NewClientError(err)
		}
	}

	// Embed the token in the request
	req.Header.Add("Authorization", "Bearer "+c.token.AccessToken)
	// And rest client revision
	req.Header.Add("x-mifasol-client-version", version.AppVersion.String())

	// Add optional body content for POST & PUT request
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}

	// Send the request
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewClientError(err)
	}

	// Is the response OK ?
	cliErr = checkStatusCode(response)
	if cliErr != nil {
		// Is the token expired ?
		if cliErr.Code() == restApiV1.InvalidTokenErrorCode {
			// Ask a new one and retry
			c.token = nil
			return c.doRequest(method, relativeUrl, contentType, body)
		}

		return nil, cliErr
	}

	// Return response
	return response, nil
}

func (c *RestClient) doGetRequest(relativeUrl string) (*http.Response, ClientError) {
	return c.doRequest("GET", relativeUrl, "", nil)
}
func (c *RestClient) doGetRequestWithBody(relativeUrl string, contentType string, body io.Reader) (*http.Response, ClientError) {
	return c.doRequest("GET", relativeUrl, contentType, body)
}
func (c *RestClient) doDeleteRequest(relativeUrl string) (*http.Response, ClientError) {
	return c.doRequest("DELETE", relativeUrl, "", nil)
}
func (c *RestClient) doPostRequest(relativeUrl string, contentType string, body io.Reader) (*http.Response, ClientError) {
	return c.doRequest("POST", relativeUrl, contentType, body)
}
func (c *RestClient) doPutRequest(relativeUrl string, contentType string, body io.Reader) (*http.Response, ClientError) {
	return c.doRequest("PUT", relativeUrl, contentType, body)
}

func checkStatusCode(response *http.Response) ClientError {

	if response.StatusCode >= 400 {
		var apiErr restApiV1.ApiError
		if err := json.NewDecoder(response.Body).Decode(&apiErr); err != nil {
			apiErr.ErrorCode = restApiV1.UnknownErrorCode
		}

		return &apiErr
	}

	return nil
}

func (c *RestClient) UserId() restApiV1.UserId {
	token, cliErr := c.GetToken()
	if cliErr != nil {
		return restApiV1.UndefinedUserId
	} else {
		return token.UserId
	}
}
