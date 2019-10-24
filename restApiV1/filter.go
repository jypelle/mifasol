package restApiV1

type ArtistOrder int64

const (
	ArtistOrderNoOrder ArtistOrder = iota
	ArtistOrderByArtistName
	ArtistOrderByUpdateTs
)

type ArtistFilter struct {
	Order  ArtistOrder
	FromTs *int64
	Name   *string
}

type AlbumOrder int64

const (
	AlbumOrderNoOrder AlbumOrder = iota
	AlbumOrderByAlbumName
	AlbumOrderByUpdateTs
)

type AlbumFilter struct {
	Order  AlbumOrder
	FromTs *int64
	Name   *string
}

type PlaylistOrder int64

const (
	PlaylistOrderNoOrder PlaylistOrder = iota
	PlaylistOrderByPlaylistName
	PlaylistOrderByUpdateTs
	PlaylistOrderByContentUpdateTs
)

type PlaylistFilter struct {
	Order          PlaylistOrder
	FromTs         *int64
	ContentFromTs  *int64
	FavoriteUserId *UserId
	FavoriteFromTs *int64
}

type SongOrder int64

const (
	SongOrderByNoOrder SongOrder = iota
	SongOrderBySongName
	SongOrderByUpdateTs
)

type SongFilter struct {
	Order   SongOrder
	FromTs  *int64
	AlbumId *string
}

type UserOrder int64

const (
	UserOrderByNoOrder UserOrder = iota
	UserOrderByUserName
	UserOrderByUpdateTs
)

type UserFilter struct {
	Order   UserOrder
	FromTs  *int64
	AdminFg *bool
}

type FavoritePlaylistOrder int64

const (
	FavoritePlaylistOrderNoOrder FavoritePlaylistOrder = iota
	FavoritePlaylistOrderByUpdateTs
)

type FavoritePlaylistFilter struct {
	Order  FavoritePlaylistOrder
	FromTs *int64
}
