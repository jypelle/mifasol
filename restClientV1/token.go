package restClientV1

import (
	"encoding/json"
	"net/http"
)

func (c *RestClient) refreshToken() ClientError {
	c.token = nil

	req, err := http.NewRequest("POST", c.getServerUrl()+"/token", nil)
	if err != nil {
		return NewClientError(err)
	}
	query := req.URL.Query()
	query.Add("grant_type", "password")
	query.Add("username", c.ClientConfig.GetUsername())
	query.Add("password", c.ClientConfig.GetPassword())
	req.URL.RawQuery = query.Encode()

	response, err := c.httpClient.Do(req)
	if err != nil {
		return NewClientError(err)
	}
	cliErr := checkStatusCode(response)
	if cliErr != nil {
		return cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&c.token); err != nil {

		return NewClientError(err)
	}

	return nil

}
