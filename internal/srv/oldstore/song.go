package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/jypelle/mifasol/internal/srv/oldentity"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func (s *OldStore) ReadSongs(externalTrn storm.Node, filter *restApiV1.SongFilter) ([]restApiV1.Song, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadSongs")
	}

	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	songEntities := []oldentity.SongEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &songEntities)
	} else if filter.AlbumId != nil {
		e = txn.Find("AlbumId", *filter.AlbumId, &songEntities)
	} else {
		e = txn.All(&songEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	songs := []restApiV1.Song{}

	for _, songEntity := range songEntities {
		if filter.FavoriteUserId != nil {
			fav, e := s.ReadFavoriteSong(txn, restApiV1.FavoriteSongId{UserId: *filter.FavoriteUserId, SongId: songEntity.Id})
			if e != nil {
				continue
			}
			if filter.FavoriteFromTs != nil {
				if songEntity.UpdateTs < *filter.FavoriteFromTs && fav.UpdateTs < *filter.FavoriteFromTs {
					continue
				}
			}
		}

		var song restApiV1.Song
		songEntity.Fill(&song)
		songs = append(songs, song)
	}

	return songs, nil
}

func (s *OldStore) ReadSong(externalTrn storm.Node, songId restApiV1.SongId) (*restApiV1.Song, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	var songEntity oldentity.SongEntity
	e = txn.One("Id", songId, &songEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, storeerror.ErrNotFound
		}
		return nil, e
	}

	var song restApiV1.Song
	songEntity.Fill(&song)

	return &song, nil
}

func (s *OldStore) ReadSongContent(song *restApiV1.Song) ([]byte, error) {

	content, err := ioutil.ReadFile(s.GetSongFileName(song))
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (s *OldStore) GetSongDirName(songId restApiV1.SongId) string {
	return filepath.Join(s.ServerConfig.GetCompleteConfigSongsDirName(), string(songId)[len(songId)-2:])
}

func (s *OldStore) getSongFileName(songEntity *oldentity.SongEntity) string {
	return filepath.Join(s.GetSongDirName(songEntity.Id), string(songEntity.Id)+songEntity.Format.Extension())
}

func (s *OldStore) GetSongFileName(song *restApiV1.Song) string {
	return filepath.Join(s.GetSongDirName(song.Id), string(song.Id)+song.Format.Extension())
}

func (s *OldStore) CreateSong(externalTrn storm.Node, songNew *restApiV1.SongNew, check bool) (*restApiV1.Song, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	// Store song
	now := time.Now().UnixNano()

	songEntity := oldentity.SongEntity{
		Id:         restApiV1.SongId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
	}
	songEntity.LoadMeta(&songNew.SongMeta)

	// Reorder artists
	songEntity.ArtistIds = tool.DeduplicateArtistId(songEntity.ArtistIds)
	e = s.sortArtistIds(txn, songEntity.ArtistIds)
	if e != nil {
		return nil, e
	}

	// Create album link
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		if check {
			// Check album id
			var albumEntity oldentity.AlbumEntity
			e = txn.One("Id", songEntity.AlbumId, &albumEntity)
			if e != nil {
				return nil, e
			}
		}
	}

	// Create artists link
	for _, artistId := range songEntity.ArtistIds {
		// Check artist id
		if check {
			var artistEntity oldentity.ArtistEntity
			e = txn.One("Id", artistId, &artistEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store artist songs
		e = txn.Save(oldentity.NewArtistSongEntity(artistId, songEntity.Id))
		if e != nil {
			return nil, e
		}

	}

	// Create song
	e = txn.Save(&songEntity)
	if e != nil {
		return nil, e
	}

	// Write song content
	e = os.MkdirAll(s.GetSongDirName(songEntity.Id), 0770)
	if e != nil {
		return nil, e
	}

	// TESTJY
	e = ioutil.WriteFile(s.getSongFileName(&songEntity), songNew.Content, 0660)
	if e != nil {
		return nil, e
	}

	// Update tags in song content
	e = s.UpdateSongContentTag(txn, &songEntity)
	if e != nil {
		// If tags not updated, delete the song file
		os.Remove(s.getSongFileName(&songEntity))
		return nil, e
	}

	// Refresh album artists
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		e = s.refreshAlbumArtistIds(txn, songEntity.AlbumId, nil)
		if e != nil {
			return nil, e
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var song restApiV1.Song
	songEntity.Fill(&song)

	return &song, nil
}

func (s *OldStore) CreateSongFromRawContent(externalTrn storm.Node, raw io.ReadCloser, lastAlbumId restApiV1.AlbumId) (*restApiV1.Song, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	var content []byte
	content, e = ioutil.ReadAll(raw)
	if e != nil {
		return nil, e
	}

	prefix := content[:4]

	var songNew *restApiV1.SongNew

	// Extract song meta from tags
	switch string(prefix) {
	case "fLaC":
		songNew, e = s.createSongNewFromFlacContent(txn, content, lastAlbumId)
	case "OggS":
		songNew, e = s.createSongNewFromOggContent(txn, content, lastAlbumId)
	default:
		songNew, e = s.createSongNewFromMp3Content(txn, content, lastAlbumId)
	}

	if e != nil {
		return nil, e
	}

	logrus.Debugf("Create song")
	var song *restApiV1.Song
	song, e = s.CreateSong(txn, songNew, false)
	if e != nil {
		return nil, e
	}

	// Add song to incoming playlist
	logrus.Debugf("Add song to incoming playlist")
	_, e = s.AddSongToPlaylist(txn, restApiV1.IncomingPlaylistId, song.Id, false)
	if e != nil {
		return nil, e
	}

	logrus.Debugf("Commit")
	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}
	logrus.Debugf("End commit")

	return song, nil
}

func (s *OldStore) UpdateSong(externalTrn storm.Node, songId restApiV1.SongId, songMeta *restApiV1.SongMeta, updateArtistMetaArtistId *restApiV1.ArtistId, check bool) (*restApiV1.Song, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	var songEntity oldentity.SongEntity
	e = txn.One("Id", songId, &songEntity)
	if e != nil {
		return nil, e
	}

	songOldArtistIds := songEntity.ArtistIds
	songOldAlbumId := songEntity.AlbumId

	songEntity.LoadMeta(songMeta)

	// Deduplicate artists
	if songMeta != nil {
		songEntity.ArtistIds = tool.DeduplicateArtistId(songEntity.ArtistIds)
	}

	// Reorder artists
	if songMeta != nil || updateArtistMetaArtistId != nil {
		e = s.sortArtistIds(txn, songEntity.ArtistIds)
		if e != nil {
			return nil, e
		}
	}

	songEntity.UpdateTs = time.Now().UnixNano()

	// Update album link
	if songOldAlbumId != songEntity.AlbumId {
		if songEntity.AlbumId != restApiV1.UnknownAlbumId {
			// Check album id
			if check {
				var albumEntity oldentity.AlbumEntity
				e = txn.One("Id", songEntity.AlbumId, &albumEntity)
				if e != nil {
					return nil, e
				}
			}
		}
	}

	artistIdsChanged := !isArtistIdsEqual(songOldArtistIds, songEntity.ArtistIds)

	// Update artists link
	if songMeta != nil && artistIdsChanged {
		e = txn.Select(q.Eq("SongId", songEntity.Id)).Delete(&oldentity.ArtistSongEntity{})
		if e != nil && e != storm.ErrNotFound {
			return nil, e
		}
		for _, artistId := range songEntity.ArtistIds {
			// Check artist id
			if check {
				var artistEntity oldentity.ArtistEntity
				e = txn.One("Id", artistId, &artistEntity)
				if e != nil {
					return nil, e
				}
			}

			// Store artist song
			e = txn.Save(oldentity.NewArtistSongEntity(artistId, songEntity.Id))
			if e != nil {
				return nil, e
			}
		}
	}

	// Update song
	e = txn.Save(&songEntity)
	if e != nil {
		return nil, e
	}

	// Update playlists link
	var playlistIds []restApiV1.PlaylistId
	playlistIds, e = s.GetPlaylistIdsFromSongId(txn, songId)
	if e != nil {
		return nil, e
	}

	for _, playlistId := range playlistIds {
		_, e = s.UpdatePlaylist(txn, playlistId, nil, false)
		if e != nil {
			return nil, e
		}
	}

	// Update tags in song content
	e = s.UpdateSongContentTag(txn, &songEntity)
	if e != nil {
		return nil, e
	}

	// Refresh album artists
	if songEntity.AlbumId != restApiV1.UnknownAlbumId && (artistIdsChanged || updateArtistMetaArtistId != nil || songEntity.AlbumId != songOldAlbumId) {
		e = s.refreshAlbumArtistIds(txn, songEntity.AlbumId, updateArtistMetaArtistId)
		if e != nil {
			return nil, e
		}
	}
	if songOldAlbumId != restApiV1.UnknownAlbumId && songEntity.AlbumId != songOldAlbumId {
		e = s.refreshAlbumArtistIds(txn, songOldAlbumId, updateArtistMetaArtistId)
		if e != nil {
			return nil, e
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var song restApiV1.Song
	songEntity.Fill(&song)

	return &song, nil
}

func (s *OldStore) updateSongAlbumArtists(externalTrn storm.Node, songId restApiV1.SongId, artistIds []restApiV1.ArtistId) error {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return e
		}
		defer txn.Rollback()
	}

	var songEntity oldentity.SongEntity
	e = txn.One("Id", songId, &songEntity)
	if e != nil {
		return e
	}

	songEntity.UpdateTs = time.Now().UnixNano()

	// Update song
	e = txn.Update(&songEntity)
	if e != nil {
		return e
	}

	/*
		// Update song album artists tag
		e = s.UpdateSongContentTag(txn,&songEntity)
		if e != nil {
			return e
		}
	*/

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return nil
}

func (s *OldStore) DeleteSong(externalTrn storm.Node, songId restApiV1.SongId) (*restApiV1.Song, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "DeleteSong")
	}

	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(true)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	deleteTs := time.Now().UnixNano()

	var songEntity oldentity.SongEntity
	e = txn.One("Id", songId, &songEntity)
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

		newSongIds := make([]restApiV1.SongId, 0)
		for _, currentSongId := range playList.SongIds {
			if currentSongId != songId {
				newSongIds = append(newSongIds, currentSongId)
			}
		}
		playList.SongIds = newSongIds
		_, e = s.UpdatePlaylist(txn, playlistId, &playList.PlaylistMeta, false)
		if e != nil {
			return nil, e
		}
	}

	// Delete artists link
	artistSongEntities := []oldentity.ArtistSongEntity{}
	e = txn.Find("SongId", songEntity.Id, &artistSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	for _, artistSongEntity := range artistSongEntities {
		txn.DeleteStruct(&artistSongEntity)
	}

	// Delete favorite song link
	favoriteSongEntities := []oldentity.FavoriteSongEntity{}
	e = txn.Find("SongId", songId, &favoriteSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}
	for _, favoriteSongEntity := range favoriteSongEntities {
		s.DeleteFavoriteSong(txn, restApiV1.FavoriteSongId{UserId: favoriteSongEntity.UserId, SongId: favoriteSongEntity.SongId})
	}

	// Delete song
	e = txn.DeleteStruct(&songEntity)
	if e != nil {
		return nil, e
	}

	// Delete song content
	e = os.Remove(s.getSongFileName(&songEntity))
	if e != nil {
		return nil, e
	}

	// Archive songId
	e = txn.Save(&oldentity.DeletedSongEntity{Id: songEntity.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Refresh album artists
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		e = s.refreshAlbumArtistIds(txn, songEntity.AlbumId, nil)
		if e != nil {
			return nil, e
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var song restApiV1.Song
	songEntity.Fill(&song)

	return &song, nil
}

func (s *OldStore) GetDeletedSongIds(externalTrn storm.Node, fromTs int64) ([]restApiV1.SongId, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedSongIds")
	}

	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	deletedSongEntities := []oldentity.DeletedSongEntity{}

	e = txn.Range("DeleteTs", fromTs, time.Now().UnixNano(), &deletedSongEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	songIds := []restApiV1.SongId{}

	for _, deletedSongEntity := range deletedSongEntities {
		songIds = append(songIds, deletedSongEntity.Id)
	}

	return songIds, nil
}

func (s *OldStore) GetSongIdsFromArtistId(externalTrn storm.Node, artistId restApiV1.ArtistId) ([]restApiV1.SongId, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	artistSongEntities := []oldentity.ArtistSongEntity{}

	e = txn.Find("ArtistId", artistId, &artistSongEntities)

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	var songIds []restApiV1.SongId

	for _, artistSongEntity := range artistSongEntities {
		songIds = append(songIds, artistSongEntity.SongId)
	}

	return songIds, nil
}

func (s *OldStore) GetSongIdsFromAlbumId(externalTrn storm.Node, albumId restApiV1.AlbumId) ([]restApiV1.SongId, error) {
	var e error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, e = s.Db.Begin(false)
		if e != nil {
			return nil, e
		}
		defer txn.Rollback()
	}

	songEntities := []oldentity.SongEntity{}

	e = txn.Find("AlbumId", albumId, &songEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	var songIds []restApiV1.SongId

	for _, songEntity := range songEntities {
		songIds = append(songIds, songEntity.Id)
	}

	return songIds, nil
}

// UpdateSongContentTag update tags in song content
func (s *OldStore) UpdateSongContentTag(externalTrn storm.Node, songEntity *oldentity.SongEntity) error {

	switch songEntity.Format {
	case restApiV1.SongFormatFlac:
		return s.updateSongContentFlacTag(externalTrn, songEntity)
	case restApiV1.SongFormatMp3:
		return s.updateSongContentMp3Tag(externalTrn, songEntity)
	case restApiV1.SongFormatOgg:
		return s.updateSongContentOggTag(externalTrn, songEntity)
	}
	return nil
}
