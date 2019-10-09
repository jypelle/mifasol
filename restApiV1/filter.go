package restApiV1

type ArtistOrder int64

const (
	ArtistOrderByArtistId ArtistOrder = iota
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
	AlbumOrderByAlbumId AlbumOrder = iota
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
	PlaylistOrderByPlaylistId PlaylistOrder = iota
	PlaylistOrderByPlaylistName
	PlaylistOrderByUpdateTs
	PlaylistOrderByContentUpdateTs
)

type PlaylistFilter struct {
	Order         PlaylistOrder
	FromTs        *int64
	ContentFromTs *int64
}

type SongOrder int64

const (
	SongOrderBySongId SongOrder = iota
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
	UserOrderByUserId UserOrder = iota
	UserOrderByUserName
	UserOrderByUpdateTs
)

type UserFilter struct {
	Order   UserOrder
	FromTs  *int64
	AdminFg *bool
}
