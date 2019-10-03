package svc

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
	"lyra/restApiV1"
	"lyra/tool"
	"strings"
	"time"
)

func (s *Service) ReadSyncReport(fromTs int64) (*restApiV1.SyncReport, error) {
	var syncReport restApiV1.SyncReport

	var err error

	var txn *badger.Txn
	txn = s.Db.NewTransaction(false)
	defer txn.Discard()

	// Sync timestamp
	syncReport.SyncTs = time.Now().UnixNano()

	// Songs
	syncReport.Songs, err = s.ReadSongs(txn, &restApiV1.SongFilter{Order: restApiV1.SongOrderByUpdateTs, FromTs: fromTs})
	if err != nil {
		logrus.Panicf("Unable to read songs: %v", err)
	}
	syncReport.DeletedSongIds, err = s.GetDeletedSongIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted song ids: %v", err)
	}

	// Albums
	syncReport.Albums, err = s.ReadAlbums(txn, &restApiV1.AlbumFilter{Order: restApiV1.AlbumOrderByUpdateTs, FromTs: fromTs})
	if err != nil {
		logrus.Panicf("Unable to read albums: %v", err)
	}
	syncReport.DeletedAlbumIds, err = s.GetDeletedAlbumIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted album ids: %v", err)
	}

	// Artists
	syncReport.Artists, err = s.ReadArtists(txn, &restApiV1.ArtistFilter{Order: restApiV1.ArtistOrderByUpdateTs, FromTs: fromTs})
	if err != nil {
		logrus.Panicf("Unable to read artists: %v", err)
	}
	syncReport.DeletedArtistIds, err = s.GetDeletedArtistIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted artist ids: %v", err)
	}

	// Playlists
	syncReport.Playlists, err = s.ReadPlaylists(txn, &restApiV1.PlaylistFilter{Order: restApiV1.PlaylistOrderByUpdateTs, FromTs: fromTs})
	if err != nil {
		logrus.Panicf("Unable to read playlists: %v", err)
	}
	syncReport.DeletedPlaylistIds, err = s.GetDeletedPlaylistIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted playlist ids: %v", err)
	}

	// Users
	syncReport.Users, err = s.ReadUsers(txn, &restApiV1.UserFilter{Order: restApiV1.UserOrderByUpdateTs, FromTs: fromTs})
	if err != nil {
		logrus.Panicf("Unable to read users: %v", err)
	}
	syncReport.DeletedUserIds, err = s.GetDeletedUserIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted user ids: %v", err)
	}

	return &syncReport, nil
}

func (s *Service) ReadFileSyncReport(fromTs int64) (*restApiV1.FileSyncReport, error) {
	var fileSyncReport restApiV1.FileSyncReport

	var err error

	var txn *badger.Txn
	txn = s.Db.NewTransaction(false)
	defer txn.Discard()

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
	fileSyncReport.Playlists, err = s.ReadPlaylists(txn, &restApiV1.PlaylistFilter{Order: restApiV1.PlaylistOrderByContentUpdateTs, FromTs: fromTs})
	if err != nil {
		logrus.Panicf("Unable to read playlists: %v", err)
	}
	fileSyncReport.DeletedPlaylistIds, err = s.GetDeletedPlaylistIds(txn, fromTs)
	if err != nil {
		logrus.Panicf("Unable to read deleted playlist ids: %v", err)
	}

	return &fileSyncReport, nil
}

func (s *Service) ReadFileSyncSongs(externalTrn *badger.Txn, fromTs int64) ([]*restApiV1.FileSyncSong, error) {
	fileSyncSongs := []*restApiV1.FileSyncSong{}

	opts := badger.DefaultIteratorOptions

	opts.Prefix = []byte(songUpdateTsSongIdPrefix)
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	it.Seek([]byte(songUpdateTsSongIdPrefix + indexTs(fromTs)))

	for ; it.Valid(); it.Next() {
		var fileSyncSong restApiV1.FileSyncSong

		key := it.Item().KeyCopy(nil)

		songId := strings.Split(string(key), ":")[2]
		song, e := s.ReadSong(txn, songId)
		if e != nil {
			return nil, e
		}

		fileSyncSong.Id = song.Id
		fileSyncSong.UpdateTs = song.UpdateTs
		if song.AlbumId == nil {
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
			album, _ := s.ReadAlbum(txn, *song.AlbumId)
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

		fileSyncSongs = append(fileSyncSongs, &fileSyncSong)

	}

	return fileSyncSongs, nil
}
