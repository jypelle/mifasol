package restClientV1

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"github.com/jypelle/mifasol/restApiV1"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

const JsonContentType = "application/json"

type RestClient struct {
	ClientConfig RestConfig
	httpClient   *http.Client
	token        *restApiV1.Token
}

func NewRestClient(clientConfig RestConfig) (*RestClient, error) {
	var rootCAPool *x509.CertPool = nil

	if clientConfig.GetServerSsl() && clientConfig.GetServerSelfSigned() {
		certPem, err := ioutil.ReadFile(clientConfig.GetCompleteConfigCertFilename())
		if err != nil {
			return nil, errors.New("Reading server certificate failed: " + err.Error())
		}

		rootCAPool = x509.NewCertPool()
		rootCAPool.AppendCertsFromPEM(certPem)
	}

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
	}

	return restClient, nil
}

func (c *RestClient) getServerUrl() string {
	if c.ClientConfig.GetServerSsl() {
		return "https://" + c.ClientConfig.GetServerHostname() + ":" + strconv.FormatInt(c.ClientConfig.GetServerPort(), 10) + "/api/v1"
	} else {
		return "http://" + c.ClientConfig.GetServerHostname() + ":" + strconv.FormatInt(c.ClientConfig.GetServerPort(), 10) + "/api/v1"
	}
}

// doRequest prepare and send an http request, managing access token renewal for expired token
func (c *RestClient) doRequest(method, relativeUrl string, contentType string, body io.Reader) (*http.Response, ClientError) {

	// Dear mifasolsrv, could you gimme a token ?
	if c.token == nil {
		cliErr := c.refreshToken()
		if cliErr != nil {
			return nil, cliErr
		}
	}

	// Prepare the request
	req, err := http.NewRequest(method, c.getServerUrl()+relativeUrl, body)
	if err != nil {
		return nil, NewClientError(err)
	}

	// Embed the token in the request
	req.Header.Add("Authorization", "Bearer "+c.token.AccessToken)

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
	cliErr := checkStatusCode(response)
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
	if c.token == nil {
		cliErr := c.refreshToken()
		if cliErr != nil {
			return "xxx"
		}
	}
	return c.token.UserId
}
