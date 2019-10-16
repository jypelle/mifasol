package restClientV1

import (
	"bytes"
	"encoding/json"
	"mifasol/restApiV1"
)

func (c *RestClient) CreateFavoritePlaylist(favoritePlayListMeta *restApiV1.FavoritePlaylistMeta) (*restApiV1.FavoritePlaylist, ClientError) {
	var favoritePlaylist *restApiV1.FavoritePlaylist

	encodedFavoritePlaylistMeta, _ := json.Marshal(favoritePlayListMeta)

	response, cliErr := c.doPostRequest("/favoritePlaylists", JsonContentType, bytes.NewBuffer(encodedFavoritePlaylistMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&favoritePlaylist); err != nil {
		return nil, NewClientError(err)
	}

	return favoritePlaylist, nil

}

func (c *RestClient) DeleteFavoritePlaylist(favoritePlaylistId restApiV1.FavoritePlaylistId) (*restApiV1.FavoritePlaylist, ClientError) {
	var favoritePlaylist *restApiV1.FavoritePlaylist

	response, cliErr := c.doDeleteRequest("/favoritePlaylists/" + favoritePlaylistId.UserId + "/" + favoritePlaylistId.PlaylistId)
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&favoritePlaylist); err != nil {
		return nil, NewClientError(err)
	}

	return favoritePlaylist, nil
}
