package restClientV1

import (
	"bytes"
	"encoding/json"
	"mifasol/restApiV1"
)

func (c *RestClient) CreatePlaylist(playListMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, ClientError) {
	var playlist *restApiV1.Playlist

	encodedPlaylistMeta, _ := json.Marshal(playListMeta)

	response, cliErr := c.doPostRequest("/playlists", JsonContentType, bytes.NewBuffer(encodedPlaylistMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&playlist); err != nil {
		return nil, NewClientError(err)
	}

	return playlist, nil

}

func (c *RestClient) UpdatePlaylist(playlistId string, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, ClientError) {
	var playlist *restApiV1.Playlist

	encodedPlaylistMeta, _ := json.Marshal(playlistMeta)

	response, cliErr := c.doPutRequest("/playlists/"+playlistId, JsonContentType, bytes.NewBuffer(encodedPlaylistMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&playlist); err != nil {
		return nil, NewClientError(err)
	}

	return playlist, nil
}

func (c *RestClient) DeletePlaylist(playlistId string) (*restApiV1.Playlist, ClientError) {
	var playlist *restApiV1.Playlist

	response, cliErr := c.doDeleteRequest("/playlists/" + playlistId)
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&playlist); err != nil {
		return nil, NewClientError(err)
	}

	return playlist, nil
}
