package oldstore

import (
	"github.com/asdine/storm/v3"
	"github.com/jypelle/mifasol/internal/srv/legacy/oldentity"
	"github.com/jypelle/mifasol/restApiV1"
)

func (s *OldStore) createSongNewFromOggContent(externalTrn storm.Node, content []byte, lastAlbumId restApiV1.AlbumId) (*restApiV1.SongNew, error) {

	// Extract song meta from tags
	// TODO

	var artistIds []restApiV1.ArtistId

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

func (s *OldStore) updateSongContentOggTag(externalTrn storm.Node, songEntity *oldentity.SongEntity) error {
	// TODO
	return nil
}
