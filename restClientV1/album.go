package restClientV1

import (
	"bytes"
	"encoding/json"
	"mifasol/restApiV1"
)

func (c *RestClient) CreateAlbum(albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, ClientError) {
	var album *restApiV1.Album

	encodedAlbumMeta, _ := json.Marshal(albumMeta)

	response, cliErr := c.doPostRequest("/albums", JsonContentType, bytes.NewBuffer(encodedAlbumMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&album); err != nil {
		return nil, NewClientError(err)
	}

	return album, nil
}

func (c *RestClient) UpdateAlbum(albumId string, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, ClientError) {
	var album *restApiV1.Album

	encodedAlbumMeta, _ := json.Marshal(albumMeta)

	response, cliErr := c.doPutRequest("/albums/"+albumId, JsonContentType, bytes.NewBuffer(encodedAlbumMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&album); err != nil {
		return nil, NewClientError(err)
	}

	return album, nil
}

func (c *RestClient) DeleteAlbum(albumId string) (*restApiV1.Album, ClientError) {
	var album *restApiV1.Album

	response, cliErr := c.doDeleteRequest("/albums/" + albumId)
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&album); err != nil {
		return nil, NewClientError(err)
	}

	return album, nil
}
