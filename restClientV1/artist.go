package restClientV1

import (
	"bytes"
	"encoding/json"
	"github.com/jypelle/mifasol/restApiV1"
)

func (c *RestClient) CreateArtist(artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, ClientError) {
	var artist *restApiV1.Artist

	encodedArtistMeta, _ := json.Marshal(artistMeta)

	response, cliErr := c.doPostRequest("/artists", JsonContentType, bytes.NewBuffer(encodedArtistMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	err := json.NewDecoder(response.Body).Decode(&artist)
	if err != nil {
		return nil, NewClientError(err)
	}

	return artist, nil
}

func (c *RestClient) ReadArtists(artistFilter *restApiV1.ArtistFilter) ([]restApiV1.Artist, ClientError) {
	var artistList []restApiV1.Artist

	encodedArtistFilter, _ := json.Marshal(artistFilter)

	response, cliErr := c.doGetRequestWithContent("/artists", JsonContentType, bytes.NewBuffer(encodedArtistFilter))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&artistList); err != nil {
		return nil, NewClientError(err)
	}

	return artistList, nil
}

func (c *RestClient) UpdateArtist(artistId restApiV1.ArtistId, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, ClientError) {
	var artist *restApiV1.Artist

	encodedArtistMeta, _ := json.Marshal(artistMeta)

	response, cliErr := c.doPutRequest("/artists/"+string(artistId), JsonContentType, bytes.NewBuffer(encodedArtistMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	err := json.NewDecoder(response.Body).Decode(&artist)
	if err != nil {
		return nil, NewClientError(err)
	}

	return artist, nil
}

func (c *RestClient) DeleteArtist(artistId restApiV1.ArtistId) (*restApiV1.Artist, ClientError) {
	var artist *restApiV1.Artist

	response, cliErr := c.doDeleteRequest("/artists/" + string(artistId))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&artist); err != nil {
		return nil, NewClientError(err)
	}

	return artist, nil
}
