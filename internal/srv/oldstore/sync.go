package oldstore

import (
	"errors"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *OldStore) ReadSyncReport(fromTs int64) (*restApiV1.SyncReport, error) {
	var syncReport restApiV1.SyncReport

	var err error
	var txn storm.Node
	// Force db write lock to avoid sync timestamp overlap
	txn, err = s.Db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	// Sync timestamp
	syncReport.SyncTs = time.Now().UnixNano()

	// Songs
	syncReport.Songs, err = s.ReadSongs(txn, &restApiV1.SongFilter{FromTs: &fromTs})
	if err != nil {
		return nil, errors.New("Unable to read songs: " + err.Error())
	}
	syncReport.DeletedSongIds, err = s.GetDeletedSongIds(txn, fromTs)
	if err != nil {
		return nil, errors.New("Unable to read deleted song ids: " + err.Error())
	}

	// Albums
	syncReport.Albums, err = s.ReadAlbums(txn, &restApiV1.AlbumFilter{FromTs: &fromTs})
	if err != nil {
		return nil, errors.New("Unable to read albums: " + err.Error())
	}
	syncReport.DeletedAlbumIds, err = s.GetDeletedAlbumIds(txn, fromTs)
	if err != nil {
		return nil, errors.New("Unable to read deleted album ids: " + err.Error())
	}

	// Artists
	syncReport.Artists, err = s.ReadArtists(txn, &restApiV1.ArtistFilter{FromTs: &fromTs})
	if err != nil {
		return nil, errors.New("Unable to read artists: " + err.Error())
	}
	syncReport.DeletedArtistIds, err = s.GetDeletedArtistIds(txn, fromTs)
	if err != nil {
		return nil, errors.New("Unable to read deleted artist ids: " + err.Error())
	}

	// Playlists
	syncReport.Playlists, err = s.ReadPlaylists(txn, &restApiV1.PlaylistFilter{FromTs: &fromTs})
	if err != nil {
		return nil, errors.New("Unable to read playlists: " + err.Error())
	}
	syncReport.DeletedPlaylistIds, err = s.GetDeletedPlaylistIds(txn, fromTs)
	if err != nil {
		return nil, errors.New("Unable to read deleted playlist ids: " + err.Error())
	}

	// Users
	syncReport.Users, err = s.ReadUsers(txn, &restApiV1.UserFilter{FromTs: &fromTs})
	if err != nil {
		return nil, errors.New("Unable to read users: " + err.Error())
	}
	syncReport.DeletedUserIds, err = s.GetDeletedUserIds(txn, fromTs)
	if err != nil {
		return nil, errors.New("Unable to read deleted user ids: " + err.Error())
	}

	// Favorite playlists
	syncReport.FavoritePlaylists, err = s.ReadFavoritePlaylists(txn, &restApiV1.FavoritePlaylistFilter{FromTs: &fromTs})
	if err != nil {
		return nil, errors.New("Unable to read favorite playlists: " + err.Error())
	}
	syncReport.DeletedFavoritePlaylistIds, err = s.GetDeletedFavoritePlaylistIds(txn, fromTs)
	if err != nil {
		return nil, errors.New("Unable to read deleted favorite playlist ids: " + err.Error())
	}

	// Favorite songs
	syncReport.FavoriteSongs, err = s.ReadFavoriteSongs(txn, &restApiV1.FavoriteSongFilter{FromTs: &fromTs})
	if err != nil {
		return nil, errors.New("Unable to read favorite songs: " + err.Error())
	}
	syncReport.DeletedFavoriteSongIds, err = s.GetDeletedFavoriteSongIds(txn, fromTs)
	if err != nil {
		return nil, errors.New("Unable to read deleted favorite song ids: " + err.Error())
	}

	return &syncReport, nil
}

func (s *OldStore) ReadFileSyncReport(fromTs int64, userId restApiV1.UserId) (*restApiV1.FileSyncReport, error) {
	var fileSyncReport restApiV1.FileSyncReport

	var err error

	var txn storm.Node
	// Force db write lock to avoid sync timestamp overlap
	txn, err = s.Db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	// Sync timestamp
	fileSyncReport.SyncTs = time.Now().UnixNano()

	// Favorite Songs
	fileSyncReport.FileSyncSongs, err = s.ReadFileSyncSongs(txn, fromTs, userId)

	if err != nil {
		logrus.Panicf("Unable to read songs: %v", err)
	}
	fileSyncReport.DeletedSongIds, err = s.GetDeletedUserFavoriteSongIds(txn, fromTs, userId)
	if err != nil {
		logrus.Panicf("Unable to read deleted song ids: %v", err)
	}

	// Favorite Playlists
	fileSyncReport.Playlists, err = s.ReadPlaylists(txn, &restApiV1.PlaylistFilter{FavoriteFromTs: &fromTs, FavoriteUserId: &userId})
	if err != nil {
		logrus.Panicf("Unable to read playlists: %v", err)
	}
	fileSyncReport.DeletedPlaylistIds, err = s.GetDeletedUserFavoritePlaylistIds(txn, fromTs, userId)
	if err != nil {
		logrus.Panicf("Unable to read deleted playlist ids: %v", err)
	}

	return &fileSyncReport, nil
}

func (s *OldStore) ReadFileSyncSongs(externalTrn storm.Node, favoriteFromTs int64, favoriteUserId restApiV1.UserId) ([]restApiV1.FileSyncSong, error) {
	fileSyncSongs := []restApiV1.FileSyncSong{}

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

	songs, err := s.ReadSongs(txn, &restApiV1.SongFilter{FavoriteFromTs: &favoriteFromTs, FavoriteUserId: &favoriteUserId})
	if err != nil {
		return nil, err
	}

	for _, song := range songs {
		var fileSyncSong restApiV1.FileSyncSong

		fileSyncSong.Id = song.Id
		fileSyncSong.UpdateTs = song.UpdateTs
		if song.AlbumId == restApiV1.UnknownAlbumId {
			fileSyncSong.Filepath += tool.SanitizeFilename("(Unknown)") + "/"
			for ind, artistId := range song.ArtistIds {
				artist, _ := s.ReadArtist(txn, artistId)
				if ind != 0 {
					fileSyncSong.Filepath += ", "
				}
				fileSyncSong.Filepath += tool.SanitizeFilename(artist.Name)
			}
			fileSyncSong.Filepath += " - "
		} else {
			album, _ := s.ReadAlbum(txn, song.AlbumId)
			for ind, artistId := range album.ArtistIds {
				artist, _ := s.ReadArtist(txn, artistId)
				if ind != 0 {
					fileSyncSong.Filepath += ", "
				}
				fileSyncSong.Filepath += tool.SanitizeFilename(artist.Name)
			}
			if len(album.ArtistIds) > 0 {
				fileSyncSong.Filepath += " - "
			}

			fileSyncSong.Filepath += tool.SanitizeFilename(album.Name) + "/"

			if song.TrackNumber != nil {
				fileSyncSong.Filepath += fmt.Sprintf("%02d - ", *song.TrackNumber)
			}
		}
		fileSyncSong.Filepath += tool.SanitizeFilename(song.Name) + song.Format.Extension()

		fileSyncSongs = append(fileSyncSongs, fileSyncSong)

	}

	return fileSyncSongs, nil
}
