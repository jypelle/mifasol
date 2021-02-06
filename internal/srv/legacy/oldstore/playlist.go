package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"time"
)

func (s *OldStore) ReadPlaylists(externalTrn storm.Node, filter *restApiV1.PlaylistFilter) ([]restApiV1.Playlist, error) {
	if s.ServerConfig.DebugMode {
		defer tool.TimeTrack(time.Now(), "ReadPlaylists")
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

	playlistEntities := []oldentity.PlaylistEntity{}

	if filter.FromTs != nil {
		e = txn.Range("UpdateTs", *filter.FromTs, time.Now().UnixNano(), &playlistEntities)
	} else {
		e = txn.All(&playlistEntities)
	}

	if e != nil && e != storm.ErrNotFound {
		return nil, e
	}

	playlists := []restApiV1.Playlist{}

	for _, playlistEntity := range playlistEntities {
		if filter.FavoriteUserId != nil {
			fav, e := s.ReadFavoritePlaylist(txn, restApiV1.FavoritePlaylistId{UserId: *filter.FavoriteUserId, PlaylistId: playlistEntity.Id})
			if e != nil {
				continue
			}
			if filter.FavoriteFromTs != nil {
				if playlistEntity.ContentUpdateTs < *filter.FavoriteFromTs && fav.UpdateTs < *filter.FavoriteFromTs {
					continue
				}
			}
		}

		var playlist restApiV1.Playlist
		playlistEntity.Fill(&playlist)
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}
