package svc

import (
	"fmt"
	"strings"
)

const albumIdPrefix = "albumId:"
const albumNameAlbumIdPrefix = "albumName-albumId:"
const albumUpdateTsAlbumIdPrefix = "albumUpdateTs-albumId:"
const albumDeleteTsAlbumIdPrefix = "albumDeleteTs-albumId:"
const albumIdSongIdPrefix = "albumId-songId:"

const artistIdPrefix = "artistId:"
const artistNameArtistIdPrefix = "artistName-artistId:"
const artistUpdateTsArtistIdPrefix = "artistUpdateTs-artistId:"
const artistDeleteTsArtistIdPrefix = "artistDeleteTs-artistId:"
const artistIdSongIdPrefix = "artistId-songId:"

const playlistIdPrefix = "playlistId:"
const playlistNamePlaylistIdPrefix = "playlistName-playlistId:"
const playlistUpdateTsPlaylistIdPrefix = "playlistUpdateTs-playlistId:"
const playlistContentUpdateTsPlaylistIdPrefix = "playlistContentUpdateTs-playlistId:"
const playlistDeleteTsPlaylistIdPrefix = "playlistDeleteTs-playlistId:"

const songIdPrefix = "songId:"
const songNameSongIdPrefix = "songName-songId:"
const songUpdateTsSongIdPrefix = "songUpdateTs-songId:"
const songDeleteTsSongIdPrefix = "songDeleteTs-songId:"
const songIdPlaylistIdPrefix = "songId-playlistId:"

const userIdPrefix = "userId:"
const userNameUserIdPrefix = "userName-userId:"
const userUpdateTsUserIdPrefix = "userUpdateTs-userId:"
const userDeleteTsUserIdPrefix = "userDeleteTs-userId:"
const userIdOwnedPlaylistIdPrefix = "userId-ownedPlaylistId:"

const userIdFavoritePlaylistIdPrefix = "userId-favoritePlaylistId:"
const userIdFavoritePlaylistUpdateTsFavoritePlaylistIdPrefix = "userId-favoritePlaylistUpdateTs-favoritePlaylistId:"

/*
const userIdSongIdPrefix = "userId-songId:"
const songIdUserIdPrefix = "songId-userId:"

const userIdUpdateTsPlaylistIdPrefix = "userId-updateTs-playlistId:"
const playlistIdUserIdPrefix = "playlistId-userId:"
*/
// getAlbumIdKey generate the db key from album id
func getAlbumIdKey(albumId string) []byte {
	return []byte(albumIdPrefix + albumId)
}

// getAlbumNameAlbumIdKey generate the db key from album name and album id
func getAlbumNameAlbumIdKey(albumName string, albumId string) []byte {
	return []byte(albumNameAlbumIdPrefix + indexString(albumName) + ":" + albumId)
}

// getAlbumUpdateTsAlbumIdKey generate the db key from album update timestamp and album id
func getAlbumUpdateTsAlbumIdKey(albumUpdateTs int64, albumId string) []byte {
	return []byte(albumUpdateTsAlbumIdPrefix + indexTs(albumUpdateTs) + ":" + albumId)
}

// getAlbumDeleteTsAlbumIdKey generate the db key from album delete timestamp and album id
func getAlbumDeleteTsAlbumIdKey(albumDeleteTs int64, albumId string) []byte {
	return []byte(albumDeleteTsAlbumIdPrefix + indexTs(albumDeleteTs) + ":" + albumId)
}

// getAlbumIdSongIdKey generate the db key from album id and song id
func getAlbumIdSongIdKey(albumId string, songId string) []byte {
	return []byte(albumIdSongIdPrefix + albumId + ":" + songId)
}

// getArtistIdKey generate the db key from artist id
func getArtistIdKey(artistId string) []byte {
	return []byte(artistIdPrefix + artistId)
}

// getArtistNameArtistIdKey generate the db key from artist name and artist id
func getArtistNameArtistIdKey(artistName string, artistId string) []byte {
	return []byte(artistNameArtistIdPrefix + indexString(artistName) + ":" + artistId)
}

// getArtistUpdateTsArtistIdKey generate the db key from artist update timestamp and artist id
func getArtistUpdateTsArtistIdKey(artistUpdateTs int64, artistId string) []byte {
	return []byte(artistUpdateTsArtistIdPrefix + indexTs(artistUpdateTs) + ":" + artistId)
}

// getArtistIdSongIdKey generate the db key from album id and song id
func getArtistIdSongIdKey(artistId string, songId string) []byte {
	return []byte(artistIdSongIdPrefix + artistId + ":" + songId)
}

// getArtistDeleteTsArtistIdKey generate the db key from artist delete timestamp and artist id
func getArtistDeleteTsArtistIdKey(artistDeleteTs int64, artistId string) []byte {
	return []byte(artistDeleteTsArtistIdPrefix + indexTs(artistDeleteTs) + ":" + artistId)
}

// getPlaylistIdKey generate the db key from playlist id
func getPlaylistIdKey(playlistId string) []byte {
	return []byte(playlistIdPrefix + playlistId)
}

// getPlaylistNamePlaylistIdKey generate the db key from playlist name and playlist id
func getPlaylistNamePlaylistIdKey(playlistName string, playlistId string) []byte {
	return []byte(playlistNamePlaylistIdPrefix + indexString(playlistName) + ":" + playlistId)
}

// getPlaylistUpdateTsPlaylistIdKey generate the db key from playlist update timestamp and playlist id
func getPlaylistUpdateTsPlaylistIdKey(playlistUpdateTs int64, playlistId string) []byte {
	return []byte(playlistUpdateTsPlaylistIdPrefix + indexTs(playlistUpdateTs) + ":" + playlistId)
}

// getPlaylistContentUpdateTsPlaylistIdKey generate the db key from playlist content update timestamp and playlist id
func getPlaylistContentUpdateTsPlaylistIdKey(playlistContentUpdateTs int64, playlistId string) []byte {
	return []byte(playlistContentUpdateTsPlaylistIdPrefix + indexTs(playlistContentUpdateTs) + ":" + playlistId)
}

// getPlaylistDeleteTsPlaylistIdKey generate the db key from playlist delete timestamp and playlist id
func getPlaylistDeleteTsPlaylistIdKey(playlistDeleteTs int64, playlistId string) []byte {
	return []byte(playlistDeleteTsPlaylistIdPrefix + indexTs(playlistDeleteTs) + ":" + playlistId)
}

// getSongIdKey generate the db key from song id
func getSongIdKey(songId string) []byte {
	return []byte(songIdPrefix + songId)
}

// getSongNameSongIdKey generate the db key from song name and song id
func getSongNameSongIdKey(songName string, songId string) []byte {
	return []byte(songNameSongIdPrefix + indexString(songName) + ":" + songId)
}

// getSongUpdateTsSongIdKey generate the db key from song update timestamp and song id
func getSongUpdateTsSongIdKey(songUpdateTs int64, songId string) []byte {
	return []byte(songUpdateTsSongIdPrefix + indexTs(songUpdateTs) + ":" + songId)
}

// getSongIdPlaylistIdKey generate the db key from song id and playlist id
func getSongIdPlaylistIdKey(songId string, playlistId string) []byte {
	return []byte(songIdPlaylistIdPrefix + songId + ":" + playlistId)
}

// getSongDeleteTsSongIdKey generate the db key from song delete timestamp and song id
func getSongDeleteTsSongIdKey(songDeleteTs int64, songId string) []byte {
	return []byte(songDeleteTsSongIdPrefix + indexTs(songDeleteTs) + ":" + songId)
}

// getUserIdKey generate the db key from user id
func getUserIdKey(userId string) []byte {
	return []byte(userIdPrefix + userId)
}

// getUserNameUserIdKey generate the db key from user name and user id
func getUserNameUserIdKey(userName string, userId string) []byte {
	return []byte(userNameUserIdPrefix + indexString(userName) + ":" + userId)
}

// getUserUpdateTsUserIdKey generate the db key from user update timestamp and user id
func getUserUpdateTsUserIdKey(userUpdateTs int64, userId string) []byte {
	return []byte(userUpdateTsUserIdPrefix + indexTs(userUpdateTs) + ":" + userId)
}

// getUserDeleteTsUserIdKey generate the db key from user delete timestamp and user id
func getUserDeleteTsUserIdKey(userDeleteTs int64, userId string) []byte {
	return []byte(userDeleteTsUserIdPrefix + indexTs(userDeleteTs) + ":" + userId)
}

// getUserIdOwnedPlaylistIdKey generate the db key from user id and owned playlist id
func getUserIdOwnedPlaylistIdKey(userId string, ownedPlaylistId string) []byte {
	return []byte(userIdOwnedPlaylistIdPrefix + userId + ":" + ownedPlaylistId)
}

func normalizeString(s string) string {
	return strings.TrimSpace(strings.TrimRight(s, "\r\n\x00"))
}

func indexString(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), ":", " ")
}

func indexTs(ts int64) string {
	return fmt.Sprintf("%20d", ts)
}
