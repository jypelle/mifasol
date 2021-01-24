package store

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/internal/srv/oldentity"
	"github.com/jypelle/mifasol/internal/srv/oldstore"
	"github.com/jypelle/mifasol/restApiV1"
)

func (s *Store) OldStoreMigration(oldStore *oldstore.OldStore) error {
	txn, e := s.db.Beginx()
	if e != nil {
		return e
	}
	defer txn.Rollback()

	// Albums
	albums, e := oldStore.ReadAlbums(nil, &restApiV1.AlbumFilter{})
	if e != nil {
		return e
	}

	for _, album := range albums {
		albumEntity := entity.AlbumEntity{
			AlbumId:    album.Id,
			CreationTs: album.CreationTs,
			UpdateTs:   album.UpdateTs,
			Name:       album.Name,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	album (
			    album_id,
				creation_ts,
			    update_ts,
				name
			)
			VALUES (
			    :album_id,
				:creation_ts,
				:update_ts,
				:name
			)
	`, &albumEntity)

		if e != nil {
			return e
		}

	}

	oldDeletedAlbumEntities := []oldentity.DeletedAlbumEntity{}

	e = oldStore.Db.All(&oldDeletedAlbumEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, oldDeletedAlbumEntity := range oldDeletedAlbumEntities {
		deletedAlbumEntity := entity.DeletedAlbumEntity{
			AlbumId:  oldDeletedAlbumEntity.Id,
			DeleteTs: oldDeletedAlbumEntity.DeleteTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	deleted_album (
			    album_id,
			    delete_ts
			)
			VALUES (
			    :album_id,
				:delete_ts
			)`,
			&deletedAlbumEntity)

		if e != nil {
			return e
		}
	}

	// Artists
	artists, e := oldStore.ReadArtists(nil, &restApiV1.ArtistFilter{})
	if e != nil {
		return e
	}

	for _, artist := range artists {
		artistEntity := entity.ArtistEntity{
			ArtistId:   artist.Id,
			CreationTs: artist.CreationTs,
			UpdateTs:   artist.UpdateTs,
			Name:       artist.Name,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	artist (
			    artist_id,
				creation_ts,
			    update_ts,
				name
			)
			VALUES (
			    :artist_id,
				:creation_ts,
				:update_ts,
				:name
			)
	`, &artistEntity)

		if e != nil {
			return e
		}

	}

	oldDeletedArtistEntities := []oldentity.DeletedArtistEntity{}

	e = oldStore.Db.All(&oldDeletedArtistEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, oldDeletedArtistEntity := range oldDeletedArtistEntities {
		deletedArtistEntity := entity.DeletedArtistEntity{
			ArtistId: oldDeletedArtistEntity.Id,
			DeleteTs: oldDeletedArtistEntity.DeleteTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	deleted_artist (
			    artist_id,
			    delete_ts
			)
			VALUES (
			    :artist_id,
				:delete_ts
			)`,
			&deletedArtistEntity)

		if e != nil {
			return e
		}
	}

	// Favorite playlist
	favoritePlaylists, e := oldStore.ReadFavoritePlaylists(nil, &restApiV1.FavoritePlaylistFilter{})
	if e != nil {
		return e
	}

	for _, favoritePlaylist := range favoritePlaylists {
		favoritePlaylistEntity := entity.FavoritePlaylistEntity{
			UserId:     favoritePlaylist.Id.UserId,
			PlaylistId: favoritePlaylist.Id.PlaylistId,
			UpdateTs:   favoritePlaylist.UpdateTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	favorite_playlist (
			    user_id,
				playlist_id,
			    update_ts
			)
			VALUES (
			    :user_id,
				:playlist_id,
			    :update_ts
			)
	`, &favoritePlaylistEntity)

		if e != nil {
			return e
		}

	}

	oldDeletedFavoritePlaylistEntities := []oldentity.DeletedFavoritePlaylistEntity{}

	e = oldStore.Db.All(&oldDeletedFavoritePlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, oldDeletedFavoritePlaylistEntity := range oldDeletedFavoritePlaylistEntities {
		deletedFavoritePlaylistEntity := entity.DeletedFavoritePlaylistEntity{
			UserId:     oldDeletedFavoritePlaylistEntity.UserId,
			PlaylistId: oldDeletedFavoritePlaylistEntity.PlaylistId,
			DeleteTs:   oldDeletedFavoritePlaylistEntity.DeleteTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	deleted_favorite_playlist (
			    user_id,
				playlist_id,
			    delete_ts
			)
			VALUES (
			    :user_id,
				:playlist_id,
				:delete_ts
			)`,
			&deletedFavoritePlaylistEntity)

		if e != nil {
			return e
		}
	}

	// Favorite song
	favoriteSongs, e := oldStore.ReadFavoriteSongs(nil, &restApiV1.FavoriteSongFilter{})
	if e != nil {
		return e
	}

	for _, favoriteSong := range favoriteSongs {
		favoriteSongEntity := entity.FavoriteSongEntity{
			UserId:   favoriteSong.Id.UserId,
			SongId:   favoriteSong.Id.SongId,
			UpdateTs: favoriteSong.UpdateTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	favorite_song (
			    user_id,
				song_id,
			    update_ts
			)
			VALUES (
			    :user_id,
				:song_id,
			    :update_ts
			)
	`, &favoriteSongEntity)

		if e != nil {
			return e
		}

	}

	oldDeletedFavoriteSongEntities := []oldentity.DeletedFavoriteSongEntity{}

	e = oldStore.Db.All(&oldDeletedFavoriteSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, oldDeletedFavoriteSongEntity := range oldDeletedFavoriteSongEntities {
		deletedFavoriteSongEntity := entity.DeletedFavoriteSongEntity{
			UserId:   oldDeletedFavoriteSongEntity.UserId,
			SongId:   oldDeletedFavoriteSongEntity.SongId,
			DeleteTs: oldDeletedFavoriteSongEntity.DeleteTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	deleted_favorite_song (
			    user_id,
				song_id,
			    delete_ts
			)
			VALUES (
			    :user_id,
				:song_id,
				:delete_ts
			)`,
			&deletedFavoriteSongEntity)

		if e != nil {
			return e
		}
	}

	// Playlist
	playlists, e := oldStore.ReadPlaylists(nil, &restApiV1.PlaylistFilter{})
	if e != nil {
		return e
	}

	for _, playlist := range playlists {
		playlistEntity := entity.PlaylistEntity{
			PlaylistId:      playlist.Id,
			CreationTs:      playlist.CreationTs,
			UpdateTs:        playlist.UpdateTs,
			ContentUpdateTs: playlist.ContentUpdateTs,
			Name:            playlist.Name,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	playlist (
			    playlist_id,
				creation_ts,
				update_ts,
			    content_update_ts,
			    name
			)
			VALUES (
			    :playlist_id,
				:creation_ts,
				:update_ts,
			    :content_update_ts,
			    :name
			)`,
			&playlistEntity,
		)

		if e != nil {
			return e
		}

		for position, songId := range playlist.SongIds {

			playlistSongEntity := entity.PlaylistSongEntity{
				PlaylistId: playlist.Id,
				Position:   int64(position),
				SongId:     songId,
			}

			_, e = txn.NamedExec(`
			INSERT INTO	playlist_song (
				playlist_id,
				position,
				song_id
			)
			VALUES (
				:playlist_id,
				:position,
				:song_id
			)`,
				&playlistSongEntity)

			if e != nil {
				return e
			}
		}

		for _, userId := range playlist.OwnerUserIds {

			playlistOwnedUserEntity := entity.PlaylistOwnedUserEntity{
				PlaylistId: playlist.Id,
				UserId:     userId,
			}

			_, e = txn.NamedExec(`
			INSERT INTO	playlist_owned_user (
				playlist_id,
				user_id
			)
			VALUES (
				:playlist_id,
				:user_id
			)`,
				&playlistOwnedUserEntity)

			if e != nil {
				return e
			}
		}

	}

	oldDeletedPlaylistEntities := []oldentity.DeletedPlaylistEntity{}

	e = oldStore.Db.All(&oldDeletedPlaylistEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, oldDeletedPlaylistEntity := range oldDeletedPlaylistEntities {
		deletedPlaylistEntity := entity.DeletedPlaylistEntity{
			PlaylistId: oldDeletedPlaylistEntity.Id,
			DeleteTs:   oldDeletedPlaylistEntity.DeleteTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	deleted_playlist (
			    playlist_id,
			    delete_ts
			)
			VALUES (
			    :playlist_id,
				:delete_ts
			)`,
			&deletedPlaylistEntity,
		)

		if e != nil {
			return e
		}
	}

	// Songs
	songs, e := oldStore.ReadSongs(nil, &restApiV1.SongFilter{})
	if e != nil {
		return e
	}

	for _, song := range songs {
		songEntity := entity.SongEntity{
			SongId:     song.Id,
			CreationTs: song.CreationTs,
			UpdateTs:   song.UpdateTs,
			Name:       song.Name,
			Format:     song.Format,
			Size:       song.Size,
			BitDepth:   song.BitDepth,
			AlbumId:    song.AlbumId,
			ExplicitFg: song.ExplicitFg,
		}
		if song.PublicationYear != nil {
			songEntity.PublicationYear.Int64 = *song.PublicationYear
			songEntity.PublicationYear.Valid = true
		}
		if song.TrackNumber != nil {
			songEntity.TrackNumber.Int64 = *song.TrackNumber
			songEntity.TrackNumber.Valid = true
		}

		_, e = txn.NamedExec(`
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

		if e != nil {
			return e
		}

		for _, artistId := range song.ArtistIds {
			artistSongEntity := entity.ArtistSongEntity{
				ArtistId: artistId,
				SongId:   song.Id,
			}

			_, e = txn.NamedExec(`
			INSERT INTO	artist_song (
			    artist_id,
			    song_id
			)
			VALUES (
			    :artist_id,
				:song_id
			)`,
				&artistSongEntity)

			if e != nil {
				return e
			}
		}
	}

	oldDeletedSongEntities := []oldentity.DeletedSongEntity{}

	e = oldStore.Db.All(&oldDeletedSongEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, oldDeletedSongEntity := range oldDeletedSongEntities {
		deletedSongEntity := entity.DeletedSongEntity{
			SongId:   oldDeletedSongEntity.Id,
			DeleteTs: oldDeletedSongEntity.DeleteTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	deleted_song (
			    song_id,
			    delete_ts
			)
			VALUES (
			    :song_id,
				:delete_ts
			)`,
			&deletedSongEntity)

		if e != nil {
			return e
		}
	}

	// Users
	users, e := oldStore.ReadUsers(nil, &restApiV1.UserFilter{})
	if e != nil {
		return e
	}

	for _, user := range users {
		userEntity := entity.UserEntity{
			UserId:         user.Id,
			CreationTs:     user.CreationTs,
			UpdateTs:       user.UpdateTs,
			Name:           user.Name,
			HideExplicitFg: user.HideExplicitFg,
			AdminFg:        user.AdminFg,
			Password:       user.Password,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	user (
			    user_id,
				creation_ts,
			    update_ts,
				name,
				hide_explicit_fg,
				admin_fg,
				password
			)
			VALUES (
			    :user_id,
				:creation_ts,
				:update_ts,
				:name,
				:hide_explicit_fg,
				:admin_fg,
				:password
			)
	`, &userEntity)

		if e != nil {
			return e
		}

	}

	oldDeletedUserEntities := []oldentity.DeletedUserEntity{}

	e = oldStore.Db.All(&oldDeletedUserEntities)
	if e != nil && e != storm.ErrNotFound {
		return e
	}

	for _, oldDeletedUserEntity := range oldDeletedUserEntities {
		deletedUserEntity := entity.DeletedUserEntity{
			UserId:   oldDeletedUserEntity.Id,
			DeleteTs: oldDeletedUserEntity.DeleteTs,
		}

		_, e = txn.NamedExec(`
			INSERT INTO	deleted_user (
			    user_id,
			    delete_ts
			)
			VALUES (
			    :user_id,
				:delete_ts
			)`,
			&deletedUserEntity)

		if e != nil {
			return e
		}
	}

	e = txn.Commit()
	if e != nil {
		return e
	}

	return nil
}
