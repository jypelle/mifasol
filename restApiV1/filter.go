package restApiV1

type ArtistFilterOrderBy string

const ArtistFilterOrderByName ArtistFilterOrderBy = "name"

type ArtistFilter struct {
	FromTs  *int64
	Name    *string
	SongId  *SongId
	OrderBy *ArtistFilterOrderBy
}

type AlbumFilterOrderBy string

const AlbumFilterOrderByName AlbumFilterOrderBy = "name"

type AlbumFilter struct {
	FromTs  *int64
	Name    *string
	OrderBy *AlbumFilterOrderBy
}

type PlaylistFilterOrderBy string

const PlaylistFilterOrderByName PlaylistFilterOrderBy = "name"

type PlaylistFilter struct {
	FromTs         *int64
	FavoriteUserId *UserId
	FavoriteFromTs *int64
	OrderBy        *PlaylistFilterOrderBy
}

type SongFilterOrderBy string

const SongFilterOrderByName SongFilterOrderBy = "name"

type SongFilter struct {
	FromTs   *int64
	AlbumId  *AlbumId
	ArtistId *ArtistId
	Favorite *SongFilterFavorite
	OrderBy  *SongFilterOrderBy
}

type SongFilterFavorite struct {
	UserId UserId
	FromTs int64
}

type UserFilter struct {
	FromTs  *int64
	AdminFg *bool
}

type FavoritePlaylistFilter struct {
	FromTs     *int64
	UserId     *UserId
	PlaylistId *PlaylistId
}

type FavoriteSongFilter struct {
	FromTs *int64
}
