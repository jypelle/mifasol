package svc

import (
	"github.com/asdine/storm"
	"lyra/restApiV1"
)

func (s *Service) createSongNewFromOggContent(externalTrn storm.Node, content []byte, lastAlbumId *string) (*restApiV1.SongNew, error) {

	// Extract song meta from tags
	// TODO

	var artistIds []string

	// Check available transaction
	txn := externalTrn
	var err error
	if txn == nil {
		txn, err = s.Db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	songNew := &restApiV1.SongNew{
		SongMeta: restApiV1.SongMeta{
			Name:            "Unknown",
			Format:          restApiV1.SongFormatOgg,
			Size:            int64(len(content)),
			PublicationYear: nil,
			AlbumId:         "",
			TrackNumber:     nil,
			ArtistIds:       artistIds,
		},
		Content: content,
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return songNew, nil
}

func (s *Service) updateSongContentOggTag(externalTrn storm.Node, song *restApiV1.Song) error {
	// TODO
	return nil
}
