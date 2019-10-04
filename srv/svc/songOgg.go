package svc

import (
	"github.com/dgraph-io/badger"
	"lyra/restApiV1"
)

func (s *Service) createSongNewFromOggContent(externalTrn *badger.Txn, content []byte, lastAlbumId *string) (*restApiV1.SongNew, error) {

	// Extract song meta from tags
	// TODO

	var artistIds []string

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	songNew := &restApiV1.SongNew{
		SongMeta: restApiV1.SongMeta{
			Name:            "Unknown",
			Format:          restApiV1.SongFormatOgg,
			PublicationYear: nil,
			AlbumId:         nil,
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

func (s *Service) updateSongContentOggTag(externalTrn *badger.Txn, song *restApiV1.Song) error {
	// TODO
	return nil
}
