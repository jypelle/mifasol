package svc

import (
	"encoding/json"
	"github.com/dgraph-io/badger"
	"io"
	"io/ioutil"
	"lyra/restApiV1"
	"lyra/tool"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (s *Service) ReadSongs(externalTrn *badger.Txn, filter *restApiV1.SongFilter) ([]*restApiV1.Song, error) {
	songs := []*restApiV1.Song{}

	opts := badger.DefaultIteratorOptions
	switch filter.Order {
	case restApiV1.SongOrderBySongName:
		opts.Prefix = []byte(songNameSongIdPrefix)
		opts.PrefetchValues = false
	case restApiV1.SongOrderByUpdateTs:
		opts.Prefix = []byte(songUpdateTsSongIdPrefix)
		opts.PrefetchValues = false
	default:
		opts.Prefix = []byte(songIdPrefix)
	}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	if filter.Order == restApiV1.SongOrderByUpdateTs {
		it.Seek([]byte(songUpdateTsSongIdPrefix + indexTs(filter.FromTs)))
	} else {
		it.Rewind()
	}

	for ; it.Valid(); it.Next() {
		var song *restApiV1.Song

		switch filter.Order {
		case restApiV1.SongOrderBySongName,
			restApiV1.SongOrderByUpdateTs:
			key := it.Item().KeyCopy(nil)

			songId := strings.Split(string(key), ":")[2]
			var e error
			song, e = s.ReadSong(txn, songId)
			if e != nil {
				return nil, e
			}
		default:
			encodedSong, e := it.Item().ValueCopy(nil)
			if e != nil {
				return nil, e
			}
			e = json.Unmarshal(encodedSong, &song)
			if e != nil {
				return nil, e
			}
		}

		songs = append(songs, song)

	}

	return songs, nil
}

func (s *Service) ReadSong(externalTrn *badger.Txn, songId string) (*restApiV1.Song, error) {
	var song *restApiV1.Song

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	item, e := txn.Get(getSongIdKey(songId))
	if e != nil {
		if e == badger.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}
	encodedSong, e := item.ValueCopy(nil)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal(encodedSong, &song)
	if e != nil {
		return nil, e
	}

	return song, nil
}

func (s *Service) ReadSongContent(song *restApiV1.Song) ([]byte, error) {

	content, err := ioutil.ReadFile(s.GetSongFileName(song))
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (s *Service) GetSongDirName(songId string) string {
	return filepath.Join(s.ServerConfig.GetCompleteConfigSongsDirName(), songId[len(songId)-2:])
}

func (s *Service) GetSongFileName(song *restApiV1.Song) string {
	return filepath.Join(s.GetSongDirName(song.Id), song.Id+song.Format.Extension())
}

func (s *Service) CreateSong(externalTrn *badger.Txn, songNew *restApiV1.SongNew) (*restApiV1.Song, error) {
	var song *restApiV1.Song

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	// Create song
	now := time.Now().UnixNano()

	song = &restApiV1.Song{
		Id:         tool.CreateUlid(),
		CreationTs: now,
		UpdateTs:   now,
		SongMeta:   songNew.SongMeta,
	}

	// Reorder artists
	songNew.ArtistIds = tool.Deduplicate(songNew.ArtistIds)
	sort.Slice(songNew.ArtistIds, func(i, j int) bool {
		artistI, _ := s.ReadArtist(txn, songNew.ArtistIds[i])
		artistJ, _ := s.ReadArtist(txn, songNew.ArtistIds[j])
		return artistI.Name < artistJ.Name
	})

	// Create song name index
	e := txn.Set(getSongNameSongIdKey(song.Name, song.Id), nil)
	if e != nil {
		return nil, e
	}

	// Create updateTs Index
	e = txn.Set(getSongUpdateTsSongIdKey(song.UpdateTs, song.Id), nil)
	if e != nil {
		return nil, e
	}

	// Create album link
	if song.AlbumId != nil {
		// Check album id
		_, e := txn.Get(getAlbumIdKey(*song.AlbumId))
		if e != nil {
			return nil, e
		}

		// Store album songs
		e = txn.Set(getAlbumIdSongIdKey(*song.AlbumId, song.Id), nil)
		if e != nil {
			return nil, e
		}
	}

	// Create artists link
	for _, artistId := range songNew.ArtistIds {
		// Check artist id
		_, e := txn.Get(getArtistIdKey(artistId))
		if e != nil {
			return nil, e
		}

		// Store artist songs
		e = txn.Set(getArtistIdSongIdKey(artistId, song.Id), nil)
		if e != nil {
			return nil, e
		}
	}

	// Create song
	encodedSong, _ := json.Marshal(song)
	e = txn.Set(getSongIdKey(song.Id), encodedSong)
	if e != nil {
		return nil, e
	}

	// Write song content
	e = os.MkdirAll(s.GetSongDirName(song.Id), 0770)
	if e != nil {
		return nil, e
	}

	e = ioutil.WriteFile(s.GetSongFileName(song), songNew.Content, 0660)
	if e != nil {
		return nil, e
	}

	// Update tags in song content
	e = s.UpdateSongContentTag(txn, song)
	if e != nil {
		// If tags not updated, delete the song file
		os.Remove(s.GetSongFileName(song))
		return nil, e
	}

	// Refresh album artists
	if song.AlbumId != nil {
		e = s.refreshAlbumArtistIds(txn, *song.AlbumId, nil)
		if e != nil {
			return nil, e
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return song, nil
}

func (s *Service) CreateSongFromRawContent(externalTrn *badger.Txn, raw io.ReadCloser) (*restApiV1.Song, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	content, err := ioutil.ReadAll(raw)
	if err != nil {
		return nil, err
	}

	prefix := content[:4]

	var songNew *restApiV1.SongNew

	// Extract song meta from tags
	switch string(prefix) {
	case "fLaC":
		songNew, err = s.createSongNewFromFlacContent(txn, content)
	case "OggS":
		songNew, err = s.createSongNewFromOggContent(txn, content)
	default:
		songNew, err = s.createSongNewFromMp3Content(txn, content)
	}

	if err != nil {
		return nil, err
	}

	song, err := s.CreateSong(txn, songNew)
	if err != nil {
		return nil, err
	}

	// Add song to incoming playlist
	incomingPlayList, err := s.ReadPlaylist(txn, "00000000000000000000000000")
	if err != nil {
		return nil, err
	}
	incomingPlayList.SongIds = append(incomingPlayList.SongIds, song.Id)
	s.UpdatePlaylist(txn, incomingPlayList.Id, &incomingPlayList.PlaylistMeta)

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return song, nil
}

func (s *Service) UpdateSong(externalTrn *badger.Txn, songId string, songMeta *restApiV1.SongMeta, updateArtistMetaArtistId *string) (*restApiV1.Song, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	song, err := s.ReadSong(txn, songId)

	if err != nil {
		return nil, err
	}

	songOldName := song.Name
	songOldUpdateTs := song.UpdateTs
	songOldArtistIds := song.ArtistIds
	songOldAlbumId := song.AlbumId

	if songMeta != nil {
		song.SongMeta = *songMeta
	}

	// Reorder artists
	if songMeta != nil || updateArtistMetaArtistId != nil {
		song.ArtistIds = tool.Deduplicate(song.ArtistIds)
		sort.Slice(song.ArtistIds, func(i, j int) bool {
			artistI, _ := s.ReadArtist(txn, song.ArtistIds[i])
			artistJ, _ := s.ReadArtist(txn, song.ArtistIds[j])
			return artistI.Name < artistJ.Name
		})
	}

	song.UpdateTs = time.Now().UnixNano()

	// Update updateTs Index
	e := txn.Delete(getSongUpdateTsSongIdKey(songOldUpdateTs, song.Id))
	if e != nil {
		return nil, e
	}
	e = txn.Set(getSongUpdateTsSongIdKey(song.UpdateTs, song.Id), nil)
	if e != nil {
		return nil, e
	}

	// Update song name index
	if songOldName != song.Name {
		e = txn.Delete(getSongNameSongIdKey(songOldName, song.Id))
		if e != nil {
			return nil, e
		}
		e = txn.Set(getSongNameSongIdKey(song.Name, song.Id), nil)
		if e != nil {
			return nil, e
		}
	}

	// Update album link
	if songOldAlbumId != song.AlbumId {
		if songOldAlbumId != nil {
			e = txn.Delete(getAlbumIdSongIdKey(*songOldAlbumId, songId))
			if e != nil {
				return nil, e
			}
		}
		if song.AlbumId != nil {
			// Check album id
			_, e := txn.Get(getAlbumIdKey(*song.AlbumId))
			if e != nil {
				return nil, e
			}

			// Store album song
			e = txn.Set(getAlbumIdSongIdKey(*song.AlbumId, song.Id), nil)
			if e != nil {
				return nil, e
			}
		}
	}

	artistIdsChanged := !isArtistIdsEqual(songOldArtistIds, song.ArtistIds)

	// Update artists link
	if songMeta != nil && artistIdsChanged {
		for _, artistId := range songOldArtistIds {
			e = txn.Delete(getArtistIdSongIdKey(artistId, songId))
			if e != nil {
				return nil, e
			}
		}
		for _, artistId := range song.ArtistIds {
			// Check artist id
			_, e := txn.Get(getArtistIdKey(artistId))
			if e != nil {
				return nil, e
			}

			// Store artist song
			e = txn.Set(getArtistIdSongIdKey(artistId, song.Id), nil)
			if e != nil {
				return nil, e
			}
		}
	}

	// Update song
	encodedSong, _ := json.Marshal(song)
	e = txn.Set(getSongIdKey(song.Id), encodedSong)
	if e != nil {
		return nil, e
	}

	// Update playlists link
	playlistIds, e := s.GetPlaylistIdsFromSongId(txn, songId)
	if e != nil {
		return nil, e
	}

	for _, playlistId := range playlistIds {
		_, e = s.UpdatePlaylist(txn, playlistId, nil)
		if e != nil {
			return nil, e
		}
	}

	// Update tags in song content
	e = s.UpdateSongContentTag(txn, song)
	if e != nil {
		return nil, e
	}

	// Refresh album artists
	if song.AlbumId != nil && (artistIdsChanged || updateArtistMetaArtistId != nil || songOldAlbumId == nil || (songOldAlbumId != nil && *song.AlbumId != *songOldAlbumId)) {
		e = s.refreshAlbumArtistIds(txn, *song.AlbumId, updateArtistMetaArtistId)
		if e != nil {
			return nil, e
		}
		if songOldAlbumId != nil && *song.AlbumId != *songOldAlbumId {
			e = s.refreshAlbumArtistIds(txn, *songOldAlbumId, updateArtistMetaArtistId)
			if e != nil {
				return nil, e
			}
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return song, nil
}

func (s *Service) updateSongAlbumArtists(externalTrn *badger.Txn, songId string, artistIds []string) error {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	song, e := s.ReadSong(txn, songId)

	if e != nil {
		return e
	}

	songOldUpdateTs := song.UpdateTs

	song.UpdateTs = time.Now().UnixNano()

	// Update updateTs Index
	e = txn.Delete(getSongUpdateTsSongIdKey(songOldUpdateTs, song.Id))
	if e != nil {
		return e
	}
	e = txn.Set(getSongUpdateTsSongIdKey(song.UpdateTs, song.Id), nil)
	if e != nil {
		return e
	}

	// Update song
	encodedSong, _ := json.Marshal(song)
	e = txn.Set(getSongIdKey(song.Id), encodedSong)
	if e != nil {
		return e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return nil
}

func (s *Service) DeleteSong(externalTrn *badger.Txn, songId string) (*restApiV1.Song, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	deleteTs := time.Now().UnixNano()

	song, e := s.ReadSong(txn, songId)
	if e != nil {
		return nil, e
	}

	// Delete playlists link
	playlistIds, e := s.GetPlaylistIdsFromSongId(txn, songId)
	if e != nil {
		return nil, e
	}

	for _, playlistId := range playlistIds {
		playList, e := s.ReadPlaylist(txn, playlistId)
		if e != nil {
			return nil, e
		}

		newSongIds := make([]string, 0)
		for _, currentSongId := range playList.SongIds {
			if currentSongId != songId {
				newSongIds = append(newSongIds, currentSongId)
			}
		}
		playList.SongIds = newSongIds
		_, e = s.UpdatePlaylist(txn, playlistId, &playList.PlaylistMeta)
		if e != nil {
			return nil, e
		}
	}

	// Delete artists link
	for _, artistId := range song.ArtistIds {
		e = txn.Delete(getArtistIdSongIdKey(artistId, songId))
		if e != nil {
			return nil, e
		}
	}

	// Delete album link
	if song.AlbumId != nil {
		e = txn.Delete(getAlbumIdSongIdKey(*song.AlbumId, songId))
		if e != nil {
			return nil, e
		}
	}

	// Delete song updateTs index
	e = txn.Delete(getSongUpdateTsSongIdKey(song.UpdateTs, song.Id))
	if e != nil {
		return nil, e
	}

	// Delete song name index
	e = txn.Delete(getSongNameSongIdKey(song.Name, songId))
	if e != nil {
		return nil, e
	}

	// Delete song
	e = txn.Delete(getSongIdKey(songId))
	if e != nil {
		return nil, e
	}

	// Delete song content
	e = os.Remove(s.GetSongFileName(song))
	if e != nil {
		return nil, e
	}

	// Archive songId
	e = txn.Set(getSongDeleteTsSongIdKey(deleteTs, song.Id), nil)
	if e != nil {
		return nil, e
	}

	// Refresh album artists
	if song.AlbumId != nil {
		e = s.refreshAlbumArtistIds(txn, *song.AlbumId, nil)
		if e != nil {
			return nil, e
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return song, nil
}

func (s *Service) GetDeletedSongIds(externalTrn *badger.Txn, fromTs int64) ([]string, error) {

	songIds := []string{}

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(songDeleteTsSongIdPrefix)
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Seek([]byte(songDeleteTsSongIdPrefix + indexTs(fromTs))); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		songId := strings.Split(string(key), ":")[2]

		songIds = append(songIds, songId)

	}

	return songIds, nil
}

func (s *Service) GetSongIdsFromArtistId(externalTrn *badger.Txn, artistId string) ([]string, error) {

	var songIds []string

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(artistIdSongIdPrefix + artistId + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		songId := strings.Split(string(key), ":")[2]

		songIds = append(songIds, songId)

	}

	return songIds, nil

}

func (s *Service) GetSongIdsFromAlbumId(externalTrn *badger.Txn, albumId string) ([]string, error) {

	var songIds []string

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(albumIdSongIdPrefix + albumId + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		songId := strings.Split(string(key), ":")[2]

		songIds = append(songIds, songId)

	}

	return songIds, nil

}

// UpdateSongContentTag update tags in song content
func (s *Service) UpdateSongContentTag(externalTrn *badger.Txn, song *restApiV1.Song) error {

	switch song.Format {
	case restApiV1.SongFormatFlac:
		return s.updateSongContentFlacTag(externalTrn, song)
	case restApiV1.SongFormatMp3:
		return s.updateSongContentMp3Tag(externalTrn, song)
	case restApiV1.SongFormatOgg:
		return s.updateSongContentOggTag(externalTrn, song)

	}
	return nil
}

func (s *Service) getAlbumIdFromAlbumName(externalTrn *badger.Txn, albumName string) (*string, error) {
	var albumId *string

	if albumName != "" {

		// Check available transaction
		txn := externalTrn
		if txn == nil {
			txn = s.Db.NewTransaction(true)
			defer txn.Discard()
		}

		albumIds, err := s.GetAlbumIdsByName(txn, albumName)
		if err != nil {
			return nil, err
		}
		if len(albumIds) > 0 {
			// Link the song to an existing album
			albumId = &albumIds[0]
		} else {
			// Create the album before linking it to the song
			album, err := s.CreateAlbum(txn, &restApiV1.AlbumMeta{Name: albumName})
			if err != nil {
				return nil, err
			}
			albumId = &album.Id
		}

		// Commit transaction
		if externalTrn == nil {
			txn.Commit()
		}

	}

	return albumId, nil
}
