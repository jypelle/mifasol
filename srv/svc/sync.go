package svc

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/tool"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *Service) ReadSyncReport(fromTs int64) (*restApiV1.SyncReport, error) {
	var syncReport restApiV1.SyncReport

	var err error
	var txn storm.Node
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
		logrus.Panicf("Unable to read songs: %v", err)
	}
	syncReport.DeletedSongIds, err = s.GetDeletedSongIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted song ids: %v", err)
	}

	// Albums
	syncReport.Albums, err = s.ReadAlbums(txn, &restApiV1.AlbumFilter{FromTs: &fromTs})
	if err != nil {
		logrus.Panicf("Unable to read albums: %v", err)
	}
	syncReport.DeletedAlbumIds, err = s.GetDeletedAlbumIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted album ids: %v", err)
	}

	// Artists
	syncReport.Artists, err = s.ReadArtists(txn, &restApiV1.ArtistFilter{FromTs: &fromTs})
	if err != nil {
		logrus.Panicf("Unable to read artists: %v", err)
	}
	syncReport.DeletedArtistIds, err = s.GetDeletedArtistIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted artist ids: %v", err)
	}

	// Playlists
	syncReport.Playlists, err = s.ReadPlaylists(txn, &restApiV1.PlaylistFilter{FromTs: &fromTs})
	if err != nil {
		logrus.Panicf("Unable to read playlists: %v", err)
	}
	syncReport.DeletedPlaylistIds, err = s.GetDeletedPlaylistIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted playlist ids: %v", err)
	}

	// Users
	syncReport.Users, err = s.ReadUsers(txn, &restApiV1.UserFilter{FromTs: &fromTs})
	if err != nil {
		logrus.Panicf("Unable to read users: %v", err)
	}
	syncReport.DeletedUserIds, err = s.GetDeletedUserIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted user ids: %v", err)
	}

	// Favorite playlists
	syncReport.FavoritePlaylists, err = s.ReadFavoritePlaylists(txn, &restApiV1.FavoritePlaylistFilter{FromTs: &fromTs})
	if err != nil {
		logrus.Panicf("Unable to read playlists: %v", err)
	}
	syncReport.DeletedFavoritePlaylistIds, err = s.GetDeletedFavoritePlaylistIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted favorite playlist ids: %v", err)
	}

	return &syncReport, nil
}

func (s *Service) ReadFileSyncReport(fromTs int64, userId restApiV1.UserId) (*restApiV1.FileSyncReport, error) {
	var fileSyncReport restApiV1.FileSyncReport

	var err error

	var txn storm.Node
	txn, err = s.Db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	// Sync timestamp
	fileSyncReport.SyncTs = time.Now().UnixNano()

	// Songs
	fileSyncReport.FileSyncSongs, err = s.ReadFileSyncSongs(txn, fromTs)

	if err != nil {
		logrus.Panicf("Unable to read songs: %v", err)
	}
	fileSyncReport.DeletedSongIds, err = s.GetDeletedSongIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted song ids: %v", err)
	}

	// Playlists
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

func (s *Service) ReadFileSyncSongs(externalTrn storm.Node, fromTs int64) ([]restApiV1.FileSyncSong, error) {
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

	songs, err := s.ReadSongs(txn, &restApiV1.SongFilter{Order: restApiV1.SongOrderByUpdateTs, FromTs: &fromTs})
	if err != nil {
		return nil, err
	}

	for _, song := range songs {
		var fileSyncSong restApiV1.FileSyncSong

		fileSyncSong.Id = song.Id
		fileSyncSong.UpdateTs = song.UpdateTs
		if song.AlbumId == "" {
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
