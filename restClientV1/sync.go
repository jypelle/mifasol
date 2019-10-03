package restClientV1

import (
	"encoding/json"
	"lyra/restApiV1"
	"strconv"
)

func (c *RestClient) ReadSyncReport(fromTs int64) (*restApiV1.SyncReport, ClientError) {

	var syncReport *restApiV1.SyncReport

	response, cliErr := c.doGetRequest("/syncReport/" + strconv.FormatInt(fromTs, 10))

	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&syncReport); err != nil {
		return nil, NewClientError(err)
	}

	return syncReport, nil
}

func (c *RestClient) ReadFileSyncReport(fromTs int64) (*restApiV1.FileSyncReport, ClientError) {

	var fileSyncReport *restApiV1.FileSyncReport

	response, cliErr := c.doGetRequest("/fileSyncReport/" + strconv.FormatInt(fromTs, 10))

	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&fileSyncReport); err != nil {
		return nil, NewClientError(err)
	}

	return fileSyncReport, nil
}
