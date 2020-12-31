package svc

import "errors"

var (
	ErrDeleteArtistWithSongs = errors.New("Unable to delete an artist linked to songs")
	ErrDeleteAlbumWithSongs  = errors.New("Unable to delete an album linked to songs")
	ErrNotFound              = errors.New("Unable to find the item")
)
