package svc

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/srv/entity"
	"github.com/jypelle/mifasol/tool"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func (s *Service) ReadSongs(externalTrn storm.Node, filter *restApiV1.SongFilter) ([]restApiV1.Song, error) {
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
	}

	songEntities := []entity.SongEntity{}
	e = query.Find(&songEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	songs := []restApiV1.Song{}

	for _, songEntity := range songEntities {
		var song restApiV1.Song
		songEntity.Fill(&song)
		songs = append(songs, song)
	}

	return songs, nil
}

func (s *Service) ReadSong(externalTrn storm.Node, songId string) (*restApiV1.Song, error) {
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

	var songEntity entity.SongEntity
	e = txn.One("Id", songId, &songEntity)
	if e != nil {
		if e == storm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}

	var song restApiV1.Song
	songEntity.Fill(&song)

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

func (s *Service) getSongFileName(songEntity *entity.SongEntity) string {
	return filepath.Join(s.GetSongDirName(songEntity.Id), songEntity.Id+songEntity.Format.Extension())
}

func (s *Service) GetSongFileName(song *restApiV1.Song) string {
	return filepath.Join(s.GetSongDirName(song.Id), song.Id+song.Format.Extension())
}

func (s *Service) CreateSong(externalTrn storm.Node, songNew *restApiV1.SongNew, check bool) (*restApiV1.Song, error) {
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

	songEntity := entity.SongEntity{
		Id:         tool.CreateUlid(),
		CreationTs: now,
		UpdateTs:   now,
	}
	songEntity.LoadMeta(&songNew.SongMeta)

	// Reorder artists
	songEntity.ArtistIds = tool.Deduplicate(songEntity.ArtistIds)
	sort.Slice(songEntity.ArtistIds, func(i, j int) bool {
		artistI, _ := s.ReadArtist(txn, songEntity.ArtistIds[i])
		artistJ, _ := s.ReadArtist(txn, songEntity.ArtistIds[j])
		return artistI.Name < artistJ.Name
	})

	// Create album link
	if songEntity.AlbumId != "" {
		if check {
			// Check album id
			var albumEntity entity.AlbumEntity
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
			var artistEntity entity.ArtistEntity
			e = txn.One("Id", artistId, &artistEntity)
			if e != nil {
				return nil, e
			}
		}

		// Store artist songs
		e = txn.Save(entity.NewArtistSongEntity(artistId, songEntity.Id))
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
	if songEntity.AlbumId != "" {
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

func (s *Service) CreateSongFromRawContent(externalTrn storm.Node, raw io.ReadCloser, lastAlbumId *string) (*restApiV1.Song, error) {
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

func (s *Service) UpdateSong(externalTrn storm.Node, songId string, songMeta *restApiV1.SongMeta, updateArtistMetaArtistId *string, check bool) (*restApiV1.Song, error) {
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

	var songEntity entity.SongEntity
	e = txn.One("Id", songId, &songEntity)
	if e != nil {
		return nil, e
	}

	songOldArtistIds := songEntity.ArtistIds
	songOldAlbumId := songEntity.AlbumId

	songEntity.LoadMeta(songMeta)

	// Reorder artists
	if songMeta != nil || updateArtistMetaArtistId != nil {
		songEntity.ArtistIds = tool.Deduplicate(songEntity.ArtistIds)
		sort.Slice(songEntity.ArtistIds, func(i, j int) bool {
			artistI, _ := s.ReadArtist(txn, songEntity.ArtistIds[i])
			artistJ, _ := s.ReadArtist(txn, songEntity.ArtistIds[j])
			return artistI.Name < artistJ.Name
		})
	}

	songEntity.UpdateTs = time.Now().UnixNano()

	// Update album link
	if songOldAlbumId != songEntity.AlbumId {
		if songEntity.AlbumId != "" {
			// Check album id
			if check {
				var albumEntity entity.AlbumEntity
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
		for _, artistId := range songOldArtistIds {
			e = txn.DeleteStruct(entity.NewArtistSongEntity(artistId, songEntity.Id))
			if e != nil {
				return nil, e
			}
		}
		for _, artistId := range songEntity.ArtistIds {
			// Check artist id
			if check {
				var artistEntity entity.ArtistEntity
				e = txn.One("Id", artistId, &artistEntity)
				if e != nil {
					return nil, e
				}
			}

			// Store artist song
			e = txn.Save(entity.NewArtistSongEntity(artistId, songEntity.Id))
			if e != nil {
				return nil, e
			}
		}
	}

	// Update song
	e = txn.Update(&songEntity)
	if e != nil {
		return nil, e
	}

	// Update playlists link
	var playlistIds []string
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
	if songEntity.AlbumId != "" && (artistIdsChanged || updateArtistMetaArtistId != nil || songOldAlbumId == "" || (songOldAlbumId != "" && songEntity.AlbumId != songOldAlbumId)) {
		e = s.refreshAlbumArtistIds(txn, songEntity.AlbumId, updateArtistMetaArtistId)
		if e != nil {
			return nil, e
		}
		if songOldAlbumId != "" && songEntity.AlbumId != songOldAlbumId {
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

	var song restApiV1.Song
	songEntity.Fill(&song)

	return &song, nil
}

func (s *Service) updateSongAlbumArtists(externalTrn storm.Node, songId string, artistIds []string) error {
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

	var songEntity entity.SongEntity
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

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return nil
}

func (s *Service) DeleteSong(externalTrn storm.Node, songId string) (*restApiV1.Song, error) {
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

	var songEntity entity.SongEntity
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

		newSongIds := make([]string, 0)
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
	query := txn.Select(q.Eq("SongId", songEntity.Id))
	e = query.Delete(new(entity.ArtistSongEntity))
	if e != nil && e != storm.ErrNotFound {
		return nil, e
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
	e = txn.Save(&entity.DeletedSongEntity{Id: songEntity.Id, DeleteTs: deleteTs})
	if e != nil {
		return nil, e
	}

	// Refresh album artists
	if songEntity.AlbumId != "" {
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

func (s *Service) GetDeletedSongIds(externalTrn storm.Node, fromTs int64) ([]string, error) {
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

	query := txn.Select(q.Gte("DeleteTs", fromTs)).OrderBy("DeleteTs")

	deletedSongEntities := []entity.DeletedSongEntity{}

	e = query.Find(&deletedSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	songIds := []string{}

	for _, deletedSongEntity := range deletedSongEntities {
		songIds = append(songIds, deletedSongEntity.Id)
	}

	return songIds, nil
}

func (s *Service) GetSongIdsFromArtistId(externalTrn storm.Node, artistId string) ([]string, error) {
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

	query := txn.Select(q.Eq("ArtistId", artistId))

	artistSongEntities := []entity.ArtistSongEntity{}

	e = query.Find(&artistSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	var songIds []string

	for _, artistSongEntity := range artistSongEntities {
		songIds = append(songIds, artistSongEntity.SongId)
	}

	return songIds, nil
}

func (s *Service) GetSongIdsFromAlbumId(externalTrn storm.Node, albumId string) ([]string, error) {
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

	query := txn.Select(q.Eq("AlbumId", albumId))

	songEntities := []entity.SongEntity{}

	e = query.Find(&songEntities)
	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	var songIds []string

	for _, songEntity := range songEntities {
		songIds = append(songIds, songEntity.Id)
	}

	return songIds, nil
}

// UpdateSongContentTag update tags in song content
func (s *Service) UpdateSongContentTag(externalTrn storm.Node, songEntity *entity.SongEntity) error {

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
