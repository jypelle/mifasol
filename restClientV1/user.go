package restClientV1

import (
	"bytes"
	"encoding/json"
	"github.com/jypelle/mifasol/restApiV1"
)

func (c *RestClient) CreateUser(userMetaComplete *restApiV1.UserMetaComplete) (*restApiV1.User, ClientError) {
	var user *restApiV1.User

	encodedUserMeta, _ := json.Marshal(userMetaComplete)

	response, cliErr := c.doPostRequest("/users", JsonContentType, bytes.NewBuffer(encodedUserMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, NewClientError(err)
	}

	return user, nil

}

func (c *RestClient) UpdateUser(userId restApiV1.UserId, userMeta *restApiV1.UserMetaComplete) (*restApiV1.User, ClientError) {
	var user *restApiV1.User

	encodedUserMeta, _ := json.Marshal(userMeta)

	response, cliErr := c.doPutRequest("/users/"+string(userId), JsonContentType, bytes.NewBuffer(encodedUserMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, NewClientError(err)
	}

	return user, nil
}

func (c *RestClient) DeleteUser(userId restApiV1.UserId) (*restApiV1.User, ClientError) {
	var user *restApiV1.User

	response, cliErr := c.doDeleteRequest("/users/" + string(userId))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, NewClientError(err)
	}

	return user, nil
}
