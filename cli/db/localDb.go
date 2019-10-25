package db

import (
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"golang.org/x/text/collate"
	"sort"
)

type LocalDb struct {
	restClient *restClientV1.RestClient
	collator   *collate.Collator

	LastSyncTs int64

	Albums                  map[restApiV1.AlbumId]*restApiV1.Album
	Artists                 map[restApiV1.ArtistId]*restApiV1.Artist
	Playlists               map[restApiV1.PlaylistId]*restApiV1.Playlist
	Songs                   map[restApiV1.SongId]*restApiV1.Song
	Users                   map[restApiV1.UserId]*restApiV1.User
	UserFavoritePlaylistIds map[restApiV1.UserId]map[restApiV1.PlaylistId]struct{}
	UserFavoriteSongIds     map[restApiV1.UserId]map[restApiV1.SongId]struct{}

	OrderedAlbums    []*restApiV1.Album
	OrderedArtists   []*restApiV1.Artist
	OrderedPlaylists []*restApiV1.Playlist
	OrderedSongs     []*restApiV1.Song
	OrderedUsers     []*restApiV1.User

	UserOrderedFavoritePlaylists map[restApiV1.UserId][]*restApiV1.Playlist
	UserOrderedFavoriteSongs     map[restApiV1.UserId][]*restApiV1.Song

	AlbumOrderedSongs map[restApiV1.AlbumId][]*restApiV1.Song
	UnknownAlbumSongs []*restApiV1.Song

	ArtistOrderedSongs map[restApiV1.ArtistId][]*restApiV1.Song
	UnknownArtistSongs []*restApiV1.Song
}

func NewLocalDb(restClient *restClientV1.RestClient, collator *collate.Collator) *LocalDb {
	localDb := &LocalDb{restClient: restClient, collator: collator}

	return localDb
}

func (l *LocalDb) IsPlaylistOwnedBy(playlistId restApiV1.PlaylistId, userId restApiV1.UserId) bool {
	if playlist, ok := l.Playlists[playlistId]; ok {
		for _, ownerUserId := range playlist.OwnerUserIds {
			if ownerUserId == userId {
				return true
			}
		}
	}

	return false
}

func (l *LocalDb) Refresh() restClientV1.ClientError {

	// Retrieve library content from mifasolsrv
	syncReport, cliErr := l.restClient.ReadSyncReport(l.LastSyncTs)
	if cliErr != nil {
		return cliErr
	}

	if l.LastSyncTs == 0 {
		// Init map on first sync
		l.Songs = make(map[restApiV1.SongId]*restApiV1.Song, len(syncReport.Songs))
		l.Albums = make(map[restApiV1.AlbumId]*restApiV1.Album, len(syncReport.Albums))
		l.Artists = make(map[restApiV1.ArtistId]*restApiV1.Artist, len(syncReport.Artists))
		l.Playlists = make(map[restApiV1.PlaylistId]*restApiV1.Playlist, len(syncReport.Playlists))
		l.Users = make(map[restApiV1.UserId]*restApiV1.User, len(syncReport.Users))
		l.UserFavoritePlaylistIds = make(map[restApiV1.UserId]map[restApiV1.PlaylistId]struct{}, len(syncReport.Users))
		l.UserFavoriteSongIds = make(map[restApiV1.UserId]map[restApiV1.SongId]struct{}, len(syncReport.Users))
	} else {
		// Remove deleted items
		for _, songId := range syncReport.DeletedSongIds {
			delete(l.Songs, songId)
		}
		for _, albumId := range syncReport.DeletedAlbumIds {
			delete(l.Albums, albumId)
		}
		for _, artistId := range syncReport.DeletedArtistIds {
			delete(l.Artists, artistId)
		}
		for _, playlistId := range syncReport.DeletedPlaylistIds {
			delete(l.Playlists, playlistId)
		}
		for _, userId := range syncReport.DeletedUserIds {
			delete(l.Users, userId)
			delete(l.UserFavoritePlaylistIds, userId)
			delete(l.UserFavoriteSongIds, userId)
		}
		for _, favoritePlaylistId := range syncReport.DeletedFavoritePlaylistIds {
			if favoritePlaylistIds, ok := l.UserFavoritePlaylistIds[favoritePlaylistId.UserId]; ok {
				delete(favoritePlaylistIds, favoritePlaylistId.PlaylistId)
			}
		}
		for _, favoriteSongId := range syncReport.DeletedFavoriteSongIds {
			if favoriteSongIds, ok := l.UserFavoriteSongIds[favoriteSongId.UserId]; ok {
				delete(favoriteSongIds, favoriteSongId.SongId)
			}
		}
	}

	// Create in-memory indexes

	// Indexing songs
	for idx := range syncReport.Songs {
		song := &syncReport.Songs[idx]
		l.Songs[song.Id] = song
	}

	// Indexing albums
	for idx := range syncReport.Albums {
		album := &syncReport.Albums[idx]
		l.Albums[album.Id] = album
	}

	// Indexing artists
	for idx := range syncReport.Artists {
		artist := &syncReport.Artists[idx]
		l.Artists[artist.Id] = artist
	}

	// Indexing playlists
	for idx := range syncReport.Playlists {
		playlist := &syncReport.Playlists[idx]
		l.Playlists[playlist.Id] = playlist
	}

	// Indexing users
	for idx := range syncReport.Users {
		user := &syncReport.Users[idx]
		l.Users[user.Id] = user
		if _, ok := l.UserFavoritePlaylistIds[user.Id]; !ok {
			l.UserFavoritePlaylistIds[user.Id] = make(map[restApiV1.PlaylistId]struct{}, 2)
		}
		if _, ok := l.UserFavoriteSongIds[user.Id]; !ok {
			l.UserFavoriteSongIds[user.Id] = make(map[restApiV1.SongId]struct{}, 2)
		}
	}

	// Indexing favorite playlists
	for idx := range syncReport.FavoritePlaylists {
		favoritePlaylist := &syncReport.FavoritePlaylists[idx]
		l.UserFavoritePlaylistIds[favoritePlaylist.Id.UserId][favoritePlaylist.Id.PlaylistId] = struct{}{}
	}

	// Indexing favorite songs
	for idx := range syncReport.FavoriteSongs {
		favoriteSong := &syncReport.FavoriteSongs[idx]
		l.UserFavoriteSongIds[favoriteSong.Id.UserId][favoriteSong.Id.SongId] = struct{}{}
	}

	// OrderedSongs
	l.OrderedSongs = make([]*restApiV1.Song, 0, len(l.Songs))
	for _, song := range l.Songs {
		l.OrderedSongs = append(l.OrderedSongs, song)
	}
	sort.Slice(l.OrderedSongs, func(i, j int) bool {
		songNameCompare := l.collator.CompareString(l.OrderedSongs[i].Name, l.OrderedSongs[j].Name)
		if songNameCompare != 0 {
			return songNameCompare == -1
		} else {
			return l.OrderedSongs[i].CreationTs < l.OrderedSongs[j].CreationTs
		}
	})

	// OrderedAlbums
	l.OrderedAlbums = make([]*restApiV1.Album, 1, len(l.Albums)+1)
	for _, album := range l.Albums {
		l.OrderedAlbums = append(l.OrderedAlbums, album)
	}
	sort.Slice(l.OrderedAlbums, func(i, j int) bool {
		if l.OrderedAlbums[i] == nil {
			return true
		}
		if l.OrderedAlbums[j] == nil {
			return false
		}
		albumNameCompare := l.collator.CompareString(l.OrderedAlbums[i].Name, l.OrderedAlbums[j].Name)
		if albumNameCompare != 0 {
			return albumNameCompare == -1
		} else {
			return l.OrderedAlbums[i].CreationTs < l.OrderedAlbums[j].CreationTs
		}
	})

	// OrderedArtists
	l.OrderedArtists = make([]*restApiV1.Artist, 1, len(l.Artists)+1)
	for _, artist := range l.Artists {
		l.OrderedArtists = append(l.OrderedArtists, artist)
	}
	sort.Slice(l.OrderedArtists, func(i, j int) bool {
		if l.OrderedArtists[i] == nil {
			return true
		}
		if l.OrderedArtists[j] == nil {
			return false
		}
		artistNameCompare := l.collator.CompareString(l.OrderedArtists[i].Name, l.OrderedArtists[j].Name)
		if artistNameCompare != 0 {
			return artistNameCompare == -1
		} else {
			return l.OrderedArtists[i].CreationTs < l.OrderedArtists[j].CreationTs
		}
	})

	// AlbumOrderedSongs & ArtistOrderedSongs
	l.AlbumOrderedSongs = make(map[restApiV1.AlbumId][]*restApiV1.Song, len(l.OrderedAlbums))
	l.ArtistOrderedSongs = make(map[restApiV1.ArtistId][]*restApiV1.Song, len(l.OrderedArtists))
	l.UnknownAlbumSongs = nil
	l.UnknownArtistSongs = nil

	for _, song := range l.OrderedSongs {
		if song.AlbumId != "" {
			l.AlbumOrderedSongs[song.AlbumId] = append(l.AlbumOrderedSongs[song.AlbumId], song)
		} else {
			l.UnknownAlbumSongs = append(l.UnknownAlbumSongs, song)
		}
		if len(song.ArtistIds) > 0 {
			for _, artistId := range song.ArtistIds {
				l.ArtistOrderedSongs[artistId] = append(l.ArtistOrderedSongs[artistId], song)
			}
		} else {
			l.UnknownArtistSongs = append(l.UnknownArtistSongs, song)
		}
	}
	for _, songs := range l.AlbumOrderedSongs {
		sort.Slice(songs, func(i, j int) bool {
			if songs[i].TrackNumber != nil {
				if songs[j].TrackNumber != nil {
					if *songs[i].TrackNumber != *songs[j].TrackNumber {
						return *songs[i].TrackNumber < *songs[j].TrackNumber
					}
				} else {
					return true
				}
			} else {
				if songs[j].TrackNumber != nil {
					return false
				}
			}
			songNameCompare := l.collator.CompareString(songs[i].Name, songs[j].Name)
			if songNameCompare != 0 {
				return songNameCompare == -1
			} else {
				return songs[i].CreationTs < songs[j].CreationTs
			}
		})

	}
	sort.Slice(l.UnknownAlbumSongs, func(i, j int) bool {
		songNameCompare := l.collator.CompareString(l.UnknownAlbumSongs[i].Name, l.UnknownAlbumSongs[j].Name)
		if songNameCompare != 0 {
			return songNameCompare == -1
		} else {
			return l.UnknownAlbumSongs[i].CreationTs < l.UnknownAlbumSongs[j].CreationTs
		}
	})

	for _, songs := range l.ArtistOrderedSongs {
		sort.Slice(songs, func(i, j int) bool {
			if songs[i].AlbumId != "" {
				if songs[j].AlbumId != "" {
					if songs[i].AlbumId != songs[j].AlbumId {
						return l.collator.CompareString(l.Albums[songs[i].AlbumId].Name, l.Albums[songs[j].AlbumId].Name) == -1
					} else {
						if songs[i].TrackNumber != nil {
							if songs[j].TrackNumber != nil {
								if *songs[i].TrackNumber != *songs[j].TrackNumber {
									return *songs[i].TrackNumber < *songs[j].TrackNumber
								}
							} else {
								return true
							}
						} else {
							if songs[j].TrackNumber != nil {
								return false
							}
						}
					}
				} else {
					return false
				}
			} else {
				if songs[j].AlbumId != "" {
					return true
				}
			}

			songNameCompare := l.collator.CompareString(songs[i].Name, songs[j].Name)
			if songNameCompare != 0 {
				return songNameCompare == -1
			} else {
				return songs[i].CreationTs < songs[j].CreationTs
			}
		})
	}
	sort.Slice(l.UnknownArtistSongs, func(i, j int) bool {
		songNameCompare := l.collator.CompareString(l.UnknownArtistSongs[i].Name, l.UnknownArtistSongs[j].Name)
		if songNameCompare != 0 {
			return songNameCompare == -1
		} else {
			return l.UnknownArtistSongs[i].CreationTs < l.UnknownArtistSongs[j].CreationTs
		}
	})

	// OrderedPlaylists
	l.OrderedPlaylists = make([]*restApiV1.Playlist, 0, len(l.Playlists))
	for _, playlist := range l.Playlists {
		l.OrderedPlaylists = append(l.OrderedPlaylists, playlist)
	}
	sort.Slice(l.OrderedPlaylists, func(i, j int) bool {
		playlistNameCompare := l.collator.CompareString(l.OrderedPlaylists[i].Name, l.OrderedPlaylists[j].Name)
		if playlistNameCompare != 0 {
			return playlistNameCompare == -1
		} else {
			return l.OrderedPlaylists[i].CreationTs < l.OrderedPlaylists[j].CreationTs
		}
	})

	// OrderedUsers
	l.OrderedUsers = make([]*restApiV1.User, 0, len(l.Users))
	for _, user := range l.Users {
		l.OrderedUsers = append(l.OrderedUsers, user)
	}
	sort.Slice(l.OrderedUsers, func(i, j int) bool {
		userNameCompare := l.collator.CompareString(l.OrderedUsers[i].Name, l.OrderedUsers[j].Name)
		if userNameCompare != 0 {
			return userNameCompare == -1
		} else {
			return l.OrderedUsers[i].CreationTs < l.OrderedUsers[j].CreationTs
		}
	})

	// UserOrderedFavoritePlaylists
	l.UserOrderedFavoritePlaylists = make(map[restApiV1.UserId][]*restApiV1.Playlist, len(l.Users))
	for _, user := range l.Users {
		userOrderedPlaylists := make([]*restApiV1.Playlist, 0, len(l.UserFavoritePlaylistIds[user.Id]))
		for playlistId, _ := range l.UserFavoritePlaylistIds[user.Id] {
			userOrderedPlaylists = append(userOrderedPlaylists, l.Playlists[playlistId])
		}

		sort.Slice(userOrderedPlaylists, func(i, j int) bool {
			playlistNameCompare := l.collator.CompareString(userOrderedPlaylists[i].Name, userOrderedPlaylists[j].Name)
			if playlistNameCompare != 0 {
				return playlistNameCompare == -1
			} else {
				return userOrderedPlaylists[i].CreationTs < userOrderedPlaylists[j].CreationTs
			}
		})

		l.UserOrderedFavoritePlaylists[user.Id] = userOrderedPlaylists
	}

	// UserOrderedFavoriteSongs
	l.UserOrderedFavoriteSongs = make(map[restApiV1.UserId][]*restApiV1.Song, len(l.Users))
	for _, user := range l.Users {
		userOrderedSongs := make([]*restApiV1.Song, 0, len(l.UserFavoriteSongIds[user.Id]))
		for songId, _ := range l.UserFavoriteSongIds[user.Id] {
			userOrderedSongs = append(userOrderedSongs, l.Songs[songId])
		}

		sort.Slice(userOrderedSongs, func(i, j int) bool {
			songNameCompare := l.collator.CompareString(userOrderedSongs[i].Name, userOrderedSongs[j].Name)
			if songNameCompare != 0 {
				return songNameCompare == -1
			} else {
				return userOrderedSongs[i].CreationTs < userOrderedSongs[j].CreationTs
			}
		})

		l.UserOrderedFavoriteSongs[user.Id] = userOrderedSongs
	}

	// Remember new sync timestamp
	l.LastSyncTs = syncReport.SyncTs

	return nil
}
