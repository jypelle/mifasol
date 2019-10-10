package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"io"
	"io/ioutil"
	"lyra/restApiV1"
	"lyra/tool"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func (s *Service) ReadSongs(externalTrn storm.Node, filter *restApiV1.SongFilter) ([]restApiV1.Song, error) {
	songs := []restApiV1.Song{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var matchers []q.Matcher

	if filter.FromTs != nil {
		matchers = append(matchers, q.Gte("UpdateTs", *filter.FromTs))
	}

	if filter.AlbumId != nil {
		matchers = append(matchers, q.Eq("AlbumId", *filter.AlbumId))
	}

	query := txn.Select(matchers...)

	switch filter.Order {
	case restApiV1.SongOrderBySongName:
		query = query.OrderBy("Name")
	case restApiV1.SongOrderByUpdateTs:
		query = query.OrderBy("UpdateTs")
	default:
		query = query.OrderBy("Id")
	}

	err = query.Find(&songs)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	return songs, nil
}

func (s *Service) ReadSong(externalTrn storm.Node, songId string) (*restApiV1.Song, error) {
	var song restApiV1.Song

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	e := txn.One("Id", songId, &song)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	return &song, nil
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

func (s *Service) CreateSong(externalTrn storm.Node, songNew *restApiV1.SongNew) (*restApiV1.Song, error) {
	var song *restApiV1.Song

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
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

	// Create album link
	if song.AlbumId != "" {
		// Check album id
		var album restApiV1.Album
		e := txn.One("Id", song.AlbumId, &album)
		if e != nil {
			return nil, e
		}
	}

	// Create artists link
	for _, artistId := range songNew.ArtistIds {
		// Check artist id
		var artist restApiV1.Artist
		e := txn.One("Id", artistId, &artist)
		if e != nil {
			return nil, e
		}

		// Store artist songs
		e = txn.Save(&restApiV1.ArtistSong{ArtistSongId: restApiV1.ArtistSongId{ArtistId: artistId, SongId: song.Id}})
		if e != nil {
			return nil, e
		}

	}

	// Create song
	e := txn.Save(song)
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
	if song.AlbumId != "" {
		e = s.refreshAlbumArtistIds(txn, song.AlbumId, nil)
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

func (s *Service) CreateSongFromRawContent(externalTrn storm.Node, raw io.ReadCloser, lastAlbumId *string) (*restApiV1.Song, error) {
	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
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
		songNew, err = s.createSongNewFromFlacContent(txn, content, lastAlbumId)
	case "OggS":
		songNew, err = s.createSongNewFromOggContent(txn, content, lastAlbumId)
	default:
		songNew, err = s.createSongNewFromMp3Content(txn, content, lastAlbumId)
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

func (s *Service) UpdateSong(externalTrn storm.Node, songId string, songMeta *restApiV1.SongMeta, updateArtistMetaArtistId *string) (*restApiV1.Song, error) {
	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	song, err := s.ReadSong(txn, songId)

	if err != nil {
		return nil, err
	}

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

	// Update album link
	if songOldAlbumId != song.AlbumId {
		if song.AlbumId != "" {
			// Check album id
			var album restApiV1.Album
			e := txn.One("Id", song.AlbumId, &album)
			if e != nil {
				return nil, e
			}
		}
	}

	artistIdsChanged := !isArtistIdsEqual(songOldArtistIds, song.ArtistIds)

	// Update artists link
	if songMeta != nil && artistIdsChanged {
		for _, artistId := range songOldArtistIds {
			e := txn.DeleteStruct(&restApiV1.ArtistSong{ArtistSongId: restApiV1.ArtistSongId{ArtistId: artistId, SongId: song.Id}})
			if e != nil {
				return nil, e
			}
		}
		for _, artistId := range song.ArtistIds {
			// Check artist id
			var artist restApiV1.Artist
			e := txn.One("Id", artistId, &artist)
			if e != nil {
				return nil, e
			}

			// Store artist song
			e = txn.Save(&restApiV1.ArtistSong{ArtistSongId: restApiV1.ArtistSongId{ArtistId: artistId, SongId: song.Id}})
			if e != nil {
				return nil, e
			}
		}
	}

	// Update song
	e := txn.Update(song)
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
	if song.AlbumId != "" && (artistIdsChanged || updateArtistMetaArtistId != nil || songOldAlbumId == "" || (songOldAlbumId != "" && song.AlbumId != songOldAlbumId)) {
		e = s.refreshAlbumArtistIds(txn, song.AlbumId, updateArtistMetaArtistId)
		if e != nil {
			return nil, e
		}
		if songOldAlbumId != "" && song.AlbumId != songOldAlbumId {
			e = s.refreshAlbumArtistIds(txn, songOldAlbumId, updateArtistMetaArtistId)
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

func (s *Service) updateSongAlbumArtists(externalTrn storm.Node, songId string, artistIds []string) error {
	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return err
		}
		defer txn.Rollback()
	}

	song, err := s.ReadSong(txn, songId)

	if err != nil {
		return err
	}

	song.UpdateTs = time.Now().UnixNano()

	// Update song
	err = txn.Update(song)
	if err != nil {
		return err
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return nil
}

func (s *Service) DeleteSong(externalTrn storm.Node, songId string) (*restApiV1.Song, error) {
	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
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
	query := txn.Select(q.Eq("SongId", song.Id))
	e = query.Delete(new(restApiV1.ArtistSong))
	if e != nil {
		return nil, e
	}

	// Delete song
	e = txn.DeleteStruct(song)
	if e != nil {
		return nil, e
	}

	// Delete song content
	e = os.Remove(s.GetSongFileName(song))
	if e != nil {
		return nil, e
	}

	// Archive songId
	e = txn.Save(&restApiV1.DeletedSong{Id: song.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Refresh album artists
	if song.AlbumId != "" {
		e = s.refreshAlbumArtistIds(txn, song.AlbumId, nil)
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

func (s *Service) GetDeletedSongIds(externalTrn storm.Node, fromTs int64) ([]string, error) {

	songIds := []string{}
	deletedSongs := []restApiV1.DeletedSong{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Gte("DeleteTs", fromTs)).OrderBy("DeleteTs")

	err = query.Find(&deletedSongs)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, deletedSong := range deletedSongs {
		songIds = append(songIds, deletedSong.Id)
	}

	return songIds, nil
}

func (s *Service) GetSongIdsFromArtistId(externalTrn storm.Node, artistId string) ([]string, error) {

	var songIds []string
	artistSongs := []restApiV1.ArtistSong{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Eq("ArtistId", artistId))

	err = query.Find(&artistSongs)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, artistSong := range artistSongs {
		songIds = append(songIds, artistSong.SongId)
	}

	return songIds, nil
}

func (s *Service) GetSongIdsFromAlbumId(externalTrn storm.Node, albumId string) ([]string, error) {

	var songIds []string
	songs := []restApiV1.Song{}

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(false)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	query := txn.Select(q.Eq("AlbumId", albumId))

	err = query.Find(&songs)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	for _, song := range songs {
		songIds = append(songIds, song.Id)
	}

	return songIds, nil
}

// UpdateSongContentTag update tags in song content
func (s *Service) UpdateSongContentTag(externalTrn storm.Node, song *restApiV1.Song) error {

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
