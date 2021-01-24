package restApiV1

type ArtistOrder int64

type ArtistFilter struct {
	FromTs *int64
	Name   *string
	SongId *SongId
}

type AlbumFilter struct {
	FromTs *int64
	Name   *string
}

type PlaylistFilter struct {
	FromTs         *int64
	FavoriteUserId *UserId
	FavoriteFromTs *int64
}

type SongFilter struct {
	FromTs         *int64
	AlbumId        *AlbumId
	ArtistId       *ArtistId
	FavoriteUserId *UserId
	FavoriteFromTs *int64
}

type UserFilter struct {
	FromTs  *int64
	AdminFg *bool
}

type FavoritePlaylistFilter struct {
	FromTs *int64
}

type FavoriteSongFilter struct {
	FromTs *int64
}
