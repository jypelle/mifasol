package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/restApiV1"
)

func (s *Store) createSongNewFromOggContent(externalTrn *sqlx.Tx, content []byte, lastAlbumId restApiV1.AlbumId) (*restApiV1.SongNew, error) {

	// Extract song meta from tags
	// TODO

	var artistIds []restApiV1.ArtistId

	// Check available transaction
	var err error
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
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
			AlbumId:         restApiV1.UnknownAlbumId,
			TrackNumber:     nil,
			ExplicitFg:      false,
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

func (s *Store) updateSongContentOggTag(externalTrn *sqlx.Tx, songEntity *entity.SongEntity) error {
	// TODO
	return nil
}
