package restClientV1

import (
	"bytes"
	"encoding/json"
	"github.com/jypelle/mifasol/restApiV1"
	"io"
)

func (c *RestClient) ReadSongs(songFilter *restApiV1.SongFilter) ([]restApiV1.Song, ClientError) {
	var songList []restApiV1.Song

	encodedSongFilter, _ := json.Marshal(songFilter)

	response, cliErr := c.doGetRequestWithBody("/songs", JsonContentType, bytes.NewBuffer(encodedSongFilter))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&songList); err != nil {
		return nil, NewClientError(err)
	}

	return songList, nil
}

func (c *RestClient) ReadSong(songId restApiV1.SongId) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	response, cliErr := c.doGetRequest("/songs/" + string(songId))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}

func (c *RestClient) ReadSongContent(songId restApiV1.SongId) (io.ReadCloser, int64, ClientError) {

	response, cliErr := c.doGetRequest("/songContents/" + string(songId))
	if cliErr != nil {
		return nil, 0, cliErr
	}

	return response.Body, response.ContentLength, nil
}

func (c *RestClient) CreateSongContent(format restApiV1.SongFormat, readerSource io.Reader) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	response, cliErr := c.doPostRequest("/songContents", format.MimeType(), readerSource)
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}

func (c *RestClient) CreateSongContentForAlbum(format restApiV1.SongFormat, readerSource io.Reader, albumId restApiV1.AlbumId) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	response, cliErr := c.doPostRequest("/songContentsForAlbum/"+string(albumId), format.MimeType(), readerSource)
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}

func (c *RestClient) UpdateSong(songId restApiV1.SongId, songMeta *restApiV1.SongMeta) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	encodedSongMeta, _ := json.Marshal(songMeta)

	response, cliErr := c.doPutRequest("/songs/"+string(songId), JsonContentType, bytes.NewBuffer(encodedSongMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}

func (c *RestClient) DeleteSong(songId restApiV1.SongId) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	response, cliErr := c.doDeleteRequest("/songs/" + string(songId))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}
