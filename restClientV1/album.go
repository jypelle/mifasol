package restClientV1

import (
	"bytes"
	"encoding/json"
	"github.com/jypelle/mifasol/restApiV1"
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

func (c *RestClient) UpdateAlbum(albumId restApiV1.AlbumId, albumMeta *restApiV1.AlbumMeta) (*restApiV1.Album, ClientError) {
	var album *restApiV1.Album

	encodedAlbumMeta, _ := json.Marshal(albumMeta)

	response, cliErr := c.doPutRequest("/albums/"+string(albumId), JsonContentType, bytes.NewBuffer(encodedAlbumMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&album); err != nil {
		return nil, NewClientError(err)
	}

	return album, nil
}

func (c *RestClient) DeleteAlbum(albumId restApiV1.AlbumId) (*restApiV1.Album, ClientError) {
	var album *restApiV1.Album

	response, cliErr := c.doDeleteRequest("/albums/" + string(albumId))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&album); err != nil {
		return nil, NewClientError(err)
	}

	return album, nil
}
