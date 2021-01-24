package store

import (
	"database/sql"
	"github.com/asdine/storm/v3"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
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

func (s *Store) ReadSongs(externalTrn *sqlx.Tx, filter *restApiV1.SongFilter) ([]restApiV1.Song, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadSongs")
	}

	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	queryArgs := make(map[string]interface{})
	if filter.FromTs != nil {
		queryArgs["from_ts"] = *filter.FromTs
	}
	if filter.AlbumId != nil {
		queryArgs["album_id"] = *filter.AlbumId
	}
	if filter.AlbumId != nil {
		queryArgs["artist_id"] = *filter.ArtistId
	}
	if filter.FavoriteUserId != nil {
		queryArgs["favorite_user_id"] = *filter.FavoriteUserId
	}
	if filter.FavoriteFromTs != nil {
		queryArgs["favorite_from_ts"] = *filter.FavoriteFromTs
	}

	rows, err := txn.NamedQuery(
		`SELECT
				DISTINCT s.*
			FROM song s
			`+tool.TernStr(filter.ArtistId != nil, "JOIN artist_song as ON as.song_id = s.song_id AND s.artist_id = :artist_id ", "")+`
			`+tool.TernStr(filter.FavoriteUserId != nil || filter.FavoriteFromTs != nil, "JOIN favorite_song fs ON fs.song_id = s.song_id ", "")+`
			WHERE 1>0
			`+tool.TernStr(filter.FromTs != nil, "AND s.update_ts >= :from_ts ", "")+`
			`+tool.TernStr(filter.AlbumId != nil, "AND s.album_id = :album_id ", "")+`
			`+tool.TernStr(filter.FavoriteUserId != nil, "AND fs.user_id = :favorite_user_id ", "")+`
			`+tool.TernStr(filter.FavoriteFromTs != nil, "AND fs.update_ts >= :favorite_from_ts ", "")+`
			ORDER BY s.song_id ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	songs := []restApiV1.Song{}

	for rows.Next() {
		var songEntity entity.SongEntity
		err = rows.StructScan(&songEntity)
		if err != nil {
			return nil, err
		}

		// TODO: Need optimizations!
		// Retrieve song artists
		artistSongEntities := []entity.ArtistSongEntity{}
		err = txn.Select(&artistSongEntities, `SELECT * FROM artist_song WHERE song_id = ?`, songEntity.SongId)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, storeerror.ErrNotFound
			}
			return nil, err
		}

		var song restApiV1.Song
		songEntity.Fill(&song)
		for _, artistSongEntity := range artistSongEntities {
			song.ArtistIds = append(song.ArtistIds, artistSongEntity.ArtistId)
		}

		songs = append(songs, song)
	}

	return songs, nil
}

func (s *Store) ReadSong(externalTrn *sqlx.Tx, songId restApiV1.SongId) (*restApiV1.Song, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var songEntity entity.SongEntity

	err = txn.Get(&songEntity, "SELECT * FROM song WHERE song_id = ?", songId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	// Retrieve song artists
	artistSongEntities := []entity.ArtistSongEntity{}
	err = txn.Select(&artistSongEntities, "SELECT * FROM artist_song WHERE song_id = ?", songId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storeerror.ErrNotFound
		}
		return nil, err
	}

	var song restApiV1.Song
	songEntity.Fill(&song)
	for _, artistSongEntity := range artistSongEntities {
		song.ArtistIds = append(song.ArtistIds, artistSongEntity.ArtistId)
	}

	return &song, nil
}

func (s *Store) ReadSongContent(song *restApiV1.Song) ([]byte, error) {

	content, err := ioutil.ReadFile(s.GetSongFileName(song))
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (s *Store) GetSongDirName(songId restApiV1.SongId) string {
	return filepath.Join(s.serverConfig.GetCompleteConfigSongsDirName(), string(songId)[len(songId)-2:])
}

func (s *Store) getSongFileName(songEntity *entity.SongEntity) string {
	return filepath.Join(s.GetSongDirName(songEntity.SongId), string(songEntity.SongId)+songEntity.Format.Extension())
}

func (s *Store) GetSongFileName(song *restApiV1.Song) string {
	return filepath.Join(s.GetSongDirName(song.Id), string(song.Id)+song.Format.Extension())
}

func (s *Store) CreateSong(externalTrn *sqlx.Tx, songNew *restApiV1.SongNew, check bool) (*restApiV1.Song, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	// Store song
	now := time.Now().UnixNano()

	songEntity := entity.SongEntity{
		SongId:     restApiV1.SongId(tool.CreateUlid()),
		CreationTs: now,
		UpdateTs:   now,
	}
	songEntity.LoadMeta(&songNew.SongMeta)

	// Reorder artists
	artistIds := tool.DeduplicateArtistId(songNew.ArtistIds)
	err = s.sortArtistIds(txn, artistIds)
	if err != nil {
		return nil, err
	}

	// Create album link
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		if check {
			// Check album id
			var albumEntity oldentity.AlbumEntity
			err = txn.Get(&albumEntity, `SELECT * FROM album WHERE album_id = ?`, songEntity.AlbumId)
			if err != nil {
				return nil, err
			}
		}
	}

	// Create artists link
	for _, artistId := range artistIds {
		// Check artist id
		if check {
			var artistEntity oldentity.ArtistEntity
			err = txn.Get(&artistEntity, `SELECT * FROM artist WHERE artist_id = ?`, artistId)
			if err != nil {
				return nil, err
			}
		}

		// Store artist songs
		_, err = txn.NamedExec(`
			INSERT INTO	artist_song (
			    artist_id,
				song_id
			)
			VALUES (
			    :artist_id,
				:song_id
			)
		`, &entity.ArtistSongEntity{ArtistId: artistId, SongId: songEntity.SongId})
		if err != nil {
			return nil, err
		}
	}

	// Create song
	_, err = txn.NamedExec(`
			INSERT INTO	song (
			    song_id,
				creation_ts,
			    update_ts,
				name,
				format,
				size,
				bit_depth,
				publication_year,
				album_id,
				track_number,
				explicit_fg
			)
			VALUES (
			    :song_id,
				:creation_ts,
				:update_ts,
				:name,
				:format,
				:size,
				:bit_depth,
				:publication_year,
				:album_id,
				:track_number,
				:explicit_fg
			)`,
		&songEntity,
	)
	if err != nil {
		return nil, err
	}

	// Write song content
	err = os.MkdirAll(s.GetSongDirName(songEntity.SongId), 0770)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(s.getSongFileName(&songEntity), songNew.Content, 0660)
	if err != nil {
		return nil, err
	}

	// Update tags in song content
	err = s.UpdateSongContentTag(txn, &songEntity)
	if err != nil {
		// If tags not updated, delete the song file
		os.Remove(s.getSongFileName(&songEntity))
		return nil, err
	}

	// Refresh album artists
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		err = s.refreshAlbumArtistIds(txn, songEntity.AlbumId, nil)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	var song restApiV1.Song
	songEntity.Fill(&song)
	song.ArtistIds = artistIds

	return &song, nil
}

func (s *Store) CreateSongFromRawContent(externalTrn *sqlx.Tx, raw io.ReadCloser, lastAlbumId restApiV1.AlbumId) (*restApiV1.Song, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var content []byte
	content, err = ioutil.ReadAll(raw)
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

	logrus.Debugf("Create song")
	var song *restApiV1.Song
	song, err = s.CreateSong(txn, songNew, false)
	if err != nil {
		return nil, err
	}

	// Add song to incoming playlist
	logrus.Debugf("Add song to incoming playlist")
	_, err = s.AddSongToPlaylist(txn, restApiV1.IncomingPlaylistId, song.Id, false)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Commit")
	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}
	logrus.Debugf("End commit")

	return song, nil
}

func (s *Store) UpdateSong(externalTrn *sqlx.Tx, songId restApiV1.SongId, songMeta *restApiV1.SongMeta, updateArtistMetaArtistId *restApiV1.ArtistId, check bool) (*restApiV1.Song, error) {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var songEntity entity.SongEntity
	err = txn.Get(&songEntity, "SELECT * FROM song WHERE song_id = ?", songId)
	if err != nil {
		return nil, err
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
		err = s.sortArtistIds(txn, songEntity.ArtistIds)
		if err != nil {
			return nil, err
		}
	}

	songEntity.UpdateTs = time.Now().UnixNano()

	// Update album link
	if songOldAlbumId != songEntity.AlbumId {
		if songEntity.AlbumId != restApiV1.UnknownAlbumId {
			// Check album id
			if check {
				var albumEntity oldentity.AlbumEntity
				err = txn.One("Id", songEntity.AlbumId, &albumEntity)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	artistIdsChanged := !isArtistIdsEqual(songOldArtistIds, songEntity.ArtistIds)

	// Update artists link
	if songMeta != nil && artistIdsChanged {
		err = txn.Select(q.Eq("SongId", songEntity.Id)).Delete(&oldentity.ArtistSongEntity{})
		if err != nil && err != storm.ErrNotFound {
			return nil, err
		}
		for _, artistId := range songEntity.ArtistIds {
			// Check artist id
			if check {
				var artistEntity oldentity.ArtistEntity
				err = txn.One("Id", artistId, &artistEntity)
				if err != nil {
					return nil, err
				}
			}

			// Store artist song
			err = txn.Save(oldentity.NewArtistSongEntity(artistId, songEntity.Id))
			if err != nil {
				return nil, err
			}
		}
	}

	// Update song
	err = txn.Save(&songEntity)
	if err != nil {
		return nil, err
	}

	// Update playlists link
	var playlistIds []restApiV1.PlaylistId
	playlistIds, err = s.GetPlaylistIdsFromSongId(txn, songId)
	if err != nil {
		return nil, err
	}

	for _, playlistId := range playlistIds {
		_, err = s.UpdatePlaylist(txn, playlistId, nil, false)
		if err != nil {
			return nil, err
		}
	}

	// Update tags in song content
	err = s.UpdateSongContentTag(txn, &songEntity)
	if err != nil {
		return nil, err
	}

	// Refresh album artists
	if songEntity.AlbumId != restApiV1.UnknownAlbumId && (artistIdsChanged || updateArtistMetaArtistId != nil || songEntity.AlbumId != songOldAlbumId) {
		err = s.refreshAlbumArtistIds(txn, songEntity.AlbumId, updateArtistMetaArtistId)
		if err != nil {
			return nil, err
		}
	}
	if songOldAlbumId != restApiV1.UnknownAlbumId && songEntity.AlbumId != songOldAlbumId {
		err = s.refreshAlbumArtistIds(txn, songOldAlbumId, updateArtistMetaArtistId)
		if err != nil {
			return nil, err
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

func (s *Store) updateSongAlbumArtists(externalTrn *sqlx.Tx, songId restApiV1.SongId, artistIds []restApiV1.ArtistId) error {
	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	var songEntity oldentity.SongEntity
	err = txn.One("Id", songId, &songEntity)
	if err != nil {
		return err
	}

	songEntity.UpdateTs = time.Now().UnixNano()

	// Update song
	err = txn.Update(&songEntity)
	if err != nil {
		return err
	}

	/*
		// Update song album artists tag
		err = s.UpdateSongContentTag(txn,&songEntity)
		if err != nil {
			return err
		}
	*/

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return nil
}

func (s *Store) DeleteSong(externalTrn *sqlx.Tx, songId restApiV1.SongId) (*restApiV1.Song, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "DeleteSong")
	}

	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	deleteTs := time.Now().UnixNano()

	song, err := s.ReadSong(txn, songId)
	if err != nil {
		return nil, err
	}

	// Delete playlists link
	playlistIds, err := s.GetPlaylistIdsFromSongId(txn, songId)
	if err != nil {
		return nil, err
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
	err = txn.Find("SongId", songEntity.Id, &artistSongEntities)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}
	for _, artistSongEntity := range artistSongEntities {
		txn.DeleteStruct(&artistSongEntity)
	}

	// Delete favorite song link
	favoriteSongEntities := []oldentity.FavoriteSongEntity{}
	err = txn.Find("SongId", songId, &favoriteSongEntities)
	if err != nil && err != storm.ErrNotFound {
		return nil, err
	}
	for _, favoriteSongEntity := range favoriteSongEntities {
		s.DeleteFavoriteSong(txn, restApiV1.FavoriteSongId{UserId: favoriteSongEntity.UserId, SongId: favoriteSongEntity.SongId})
	}

	// Delete song
	err = txn.DeleteStruct(&songEntity)
	if err != nil {
		return nil, err
	}

	// Delete song content
	err = os.Remove(s.getSongFileName(&songEntity))
	if err != nil {
		return nil, err
	}

	// Archive songId
	err = txn.Save(&oldentity.DeletedSongEntity{Id: songEntity.Id, DeleteTs: deleteTs})
	if err != nil {
		return nil, err
	}

	// Refresh album artists
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		err = s.refreshAlbumArtistIds(txn, songEntity.AlbumId, nil)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return song, nil
}

func (s *Store) GetDeletedSongIds(externalTrn *sqlx.Tx, fromTs int64) ([]restApiV1.SongId, error) {
	if s.serverConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "GetDeletedSongIds")
	}

	var err error

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	queryArgs := make(map[string]interface{})
	queryArgs["from_ts"] = fromTs
	rows, err := txn.NamedQuery(
		`SELECT
				d.*
			FROM deleted_song d
			WHERE d.delete_ts >= :from_ts
			ORDER BY d.delete_ts ASC
		`,
		queryArgs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	songIds := []restApiV1.SongId{}

	for rows.Next() {
		var deletedSongEntity entity.DeletedSongEntity
		err = rows.StructScan(&deletedSongEntity)
		if err != nil {
			return nil, err
		}

		songIds = append(songIds, deletedSongEntity.SongId)
	}

	return songIds, nil
}

// UpdateSongContentTag update tags in song content
func (s *Store) UpdateSongContentTag(externalTrn *sqlx.Tx, songEntity *entity.SongEntity) error {

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
