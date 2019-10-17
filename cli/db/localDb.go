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

	Albums                map[string]*restApiV1.Album
	Artists               map[string]*restApiV1.Artist
	Playlists             map[string]*restApiV1.Playlist
	Songs                 map[string]*restApiV1.Song
	Users                 map[string]*restApiV1.User
	UserFavoritePlaylists map[string]map[string]*restApiV1.Playlist

	OrderedAlbums    []*restApiV1.Album
	OrderedArtists   []*restApiV1.Artist
	OrderedPlaylists []*restApiV1.Playlist
	OrderedSongs     []*restApiV1.Song
	OrderedUsers     []*restApiV1.User

	UserOrderedFavoritePlaylists map[string][]*restApiV1.Playlist

	AlbumOrderedSongs map[string][]*restApiV1.Song
	UnknownAlbumSongs []*restApiV1.Song

	ArtistOrderedSongs map[string][]*restApiV1.Song
	UnknownArtistSongs []*restApiV1.Song
}

func NewLocalDb(restClient *restClientV1.RestClient, collator *collate.Collator) *LocalDb {
	localDb := &LocalDb{restClient: restClient, collator: collator}

	return localDb
}

func (l *LocalDb) IsPlaylistOwnedBy(playlistId, userId string) bool {
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
		l.Songs = make(map[string]*restApiV1.Song, len(syncReport.Songs))
		l.Albums = make(map[string]*restApiV1.Album, len(syncReport.Albums))
		l.Artists = make(map[string]*restApiV1.Artist, len(syncReport.Artists))
		l.Playlists = make(map[string]*restApiV1.Playlist, len(syncReport.Playlists))
		l.Users = make(map[string]*restApiV1.User, len(syncReport.Users))
		l.UserFavoritePlaylists = make(map[string]map[string]*restApiV1.Playlist, len(syncReport.Users))
	} else {
		// Remove deleted items
		for _, favoritePlaylistId := range syncReport.DeletedFavoritePlaylistIds {
			if favoritePlaylists, ok := l.UserFavoritePlaylists[favoritePlaylistId.UserId]; ok {
				delete(favoritePlaylists, favoritePlaylistId.PlaylistId)
			}
		}
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
			delete(l.UserFavoritePlaylists, userId)
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
		l.UserFavoritePlaylists[user.Id] = make(map[string]*restApiV1.Playlist, 2)
	}

	// Indexing favorite playlists
	for idx := range syncReport.FavoritePlaylists {
		favoritePlaylist := &syncReport.FavoritePlaylists[idx]
		l.UserFavoritePlaylists[favoritePlaylist.Id.UserId][favoritePlaylist.Id.PlaylistId] = l.Playlists[favoritePlaylist.Id.PlaylistId]
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
	l.AlbumOrderedSongs = make(map[string][]*restApiV1.Song, len(l.OrderedAlbums))
	l.ArtistOrderedSongs = make(map[string][]*restApiV1.Song, len(l.OrderedArtists))
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
	l.UserOrderedFavoritePlaylists = make(map[string][]*restApiV1.Playlist, len(l.Users))
	for _, user := range l.Users {
		userOrderedPlaylists := make([]*restApiV1.Playlist, 0, len(l.UserFavoritePlaylists[user.Id]))
		for _, playlist := range l.UserFavoritePlaylists[user.Id] {
			userOrderedPlaylists = append(userOrderedPlaylists, playlist)
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

	// Remember new sync timestamp
	l.LastSyncTs = syncReport.SyncTs

	return nil
}
