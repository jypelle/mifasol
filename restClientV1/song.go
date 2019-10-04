package restClientV1

import (
	"bytes"
	"encoding/json"
	"io"
	"lyra/restApiV1"
)

func (c *RestClient) ReadSongContent(songId string) (io.ReadCloser, int64, ClientError) {

	response, cliErr := c.doGetRequest("/songContents/" + songId)
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

func (c *RestClient) CreateSongContentForAlbum(format restApiV1.SongFormat, readerSource io.Reader, albumId *string) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	var albumIdStr string
	if albumId != nil {
		albumIdStr = *albumId
	}

	response, cliErr := c.doPostRequest("/songContentsForAlbum/"+albumIdStr, format.MimeType(), readerSource)
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}

func (c *RestClient) UpdateSong(songId string, songMeta *restApiV1.SongMeta) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	encodedSongMeta, _ := json.Marshal(songMeta)

	response, cliErr := c.doPutRequest("/songs/"+songId, JsonContentType, bytes.NewBuffer(encodedSongMeta))
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}

func (c *RestClient) DeleteSong(songId string) (*restApiV1.Song, ClientError) {
	var song *restApiV1.Song

	response, cliErr := c.doDeleteRequest("/songs/" + songId)
	if cliErr != nil {
		return nil, cliErr
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&song); err != nil {
		return nil, NewClientError(err)
	}

	return song, nil
}
