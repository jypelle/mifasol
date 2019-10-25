package restClientV1

import (
	"bytes"
	"encoding/json"
	"github.com/jypelle/mifasol/restApiV1"
)

func (c *RestClient) CreateFavoriteSong(favoriteSongMeta *restApiV1.FavoriteSongMeta) (*restApiV1.FavoriteSong, ClientError) {
	var favoriteSong *restApiV1.FavoriteSong

	encodedFavoriteSongMeta, _ := json.Marshal(favoriteSongMeta)

	response, cliErr := c.doPostRequest("/favoriteSongs", JsonContentType, bytes.NewBuffer(encodedFavoriteSongMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&favoriteSong); err != nil {
		return nil, NewClientError(err)
	}

	return favoriteSong, nil

}

func (c *RestClient) DeleteFavoriteSong(favoriteSongId restApiV1.FavoriteSongId) (*restApiV1.FavoriteSong, ClientError) {
	var favoriteSong *restApiV1.FavoriteSong

	response, cliErr := c.doDeleteRequest("/favoriteSongs/" + string(favoriteSongId.UserId) + "/" + string(favoriteSongId.SongId))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&favoriteSong); err != nil {
		return nil, NewClientError(err)
	}

	return favoriteSong, nil
}
