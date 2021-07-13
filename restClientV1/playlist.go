package restClientV1

import (
	"bytes"
	"encoding/json"
	"github.com/jypelle/mifasol/restApiV1"
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

func (c *RestClient) ReadPlaylists(playlistFilter *restApiV1.PlaylistFilter) ([]restApiV1.Playlist, ClientError) {
	var playlistList []restApiV1.Playlist

	encodedPlaylistFilter, _ := json.Marshal(playlistFilter)

	response, cliErr := c.doGetRequestWithContent("/playlists", JsonContentType, bytes.NewBuffer(encodedPlaylistFilter))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&playlistList); err != nil {
		return nil, NewClientError(err)
	}

	return playlistList, nil
}

func (c *RestClient) UpdatePlaylist(playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta) (*restApiV1.Playlist, ClientError) {
	var playlist *restApiV1.Playlist

	encodedPlaylistMeta, _ := json.Marshal(playlistMeta)

	response, cliErr := c.doPutRequest("/playlists/"+string(playlistId), JsonContentType, bytes.NewBuffer(encodedPlaylistMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&playlist); err != nil {
		return nil, NewClientError(err)
	}

	return playlist, nil
}

func (c *RestClient) DeletePlaylist(playlistId restApiV1.PlaylistId) (*restApiV1.Playlist, ClientError) {
	var playlist *restApiV1.Playlist

	response, cliErr := c.doDeleteRequest("/playlists/" + string(playlistId))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&playlist); err != nil {
		return nil, NewClientError(err)
	}

	return playlist, nil
}
