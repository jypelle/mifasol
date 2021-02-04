package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
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
