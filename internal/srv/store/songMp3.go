package store

import (
	"bytes"
	"github.com/bogem/id3v2/v2"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func (s *Store) createSongNewFromMp3Content(externalTrn *sqlx.Tx, content []byte, lastAlbumId restApiV1.AlbumId) (*restApiV1.SongNew, error) {

	// Extract song meta from tags
	reader := bytes.NewReader(content)

	tag, err := id3v2.ParseReader(reader, id3v2.Options{Parse: true})
	if err != nil {
		return nil, err
	}
	defer tag.Close()

	var songNew *restApiV1.SongNew

	var bitDepth = restApiV1.SongBitDepthUnknown
	var title = ""
	var publicationYear *int64 = nil
	var albumId = restApiV1.UnknownAlbumId
	var trackNumber *int64 = nil
	var artistIds []restApiV1.ArtistId

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return nil, err
		}
		defer txn.Rollback()
	}

	// Extract title
	title = normalizeString(tag.Title())
	logrus.Debugf("Title: %s", title)

	// Extract album name
	albumName := normalizeString(tag.Album())

	// Find Album Id
	albumId, err = s.getAlbumIdFromAlbumName(txn, albumName, lastAlbumId)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Album: %s", albumName)

	// Extract track number
	if albumId != restApiV1.UnknownAlbumId {
		rawTrackNumber := strings.Split(tag.GetTextFrame(tag.CommonID("Track number/Position in set")).Text, "/")
		if len(rawTrackNumber) > 0 {
			parsedTrackNumber, _ := strconv.ParseInt(normalizeString(rawTrackNumber[0]), 10, 64)
			if parsedTrackNumber > 0 {
				trackNumber = &parsedTrackNumber
			}
		}
	}

	if trackNumber != nil {
		logrus.Debugf("Track number: %d", *trackNumber)
	}

	// Extract year
	parsedYearNumber, _ := strconv.ParseInt(normalizeString(tag.Year()), 10, 64)
	if parsedYearNumber > 0 {
		publicationYear = &parsedYearNumber
	}

	if publicationYear != nil {
		logrus.Debugf("Publication year: %d", *publicationYear)
	}

	// Extract artists
	var artistNames []string
	for _, contatArtistNames := range strings.Split(tag.Artist(), " - ") {
		artistNames = append(artistNames, strings.FieldsFunc(contatArtistNames, func(r rune) bool { return r == ',' || r == ';' || r == ';' })...)
	}

	// Find Artist IDs
	artistIds, err = s.getArtistIdsFromArtistNames(txn, artistNames)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Artists: %v", artistNames)

	songNew = &restApiV1.SongNew{
		SongMeta: restApiV1.SongMeta{
			Name:            title,
			Format:          restApiV1.SongFormatMp3,
			Size:            int64(len(content)),
			BitDepth:        bitDepth,
			PublicationYear: publicationYear,
			AlbumId:         albumId,
			TrackNumber:     trackNumber,
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

func (s *Store) updateSongContentMp3Tag(externalTrn *sqlx.Tx, songEntity *entity.SongEntity) error {
	// Extract song meta from tags
	tag, err := id3v2.Open(s.getSongFileName(songEntity.SongId, songEntity.Format), id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	// region Update tags with song meta

	// Set title
	tag.SetTitle(songEntity.Name)

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn, err = s.db.Beginx()
		if err != nil {
			return err
		}
		defer txn.Rollback()
	}

	// Set album & track number
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		album, err := s.ReadAlbum(txn, songEntity.AlbumId)
		if err != nil {
			return err
		}
		tag.SetAlbum(album.Name)

		if songEntity.TrackNumber.Valid {
			tag.AddTextFrame(tag.CommonID("Track number/Position in set"), tag.DefaultEncoding(), strconv.FormatInt(songEntity.TrackNumber.Int64, 10))
		}
	}

	// Set publication date
	if songEntity.PublicationYear.Valid {
		tag.SetYear(strconv.FormatInt(songEntity.PublicationYear.Int64, 10))
	}

	// Set artists
	artistNamesStr := ""

	artists, err := s.ReadArtists(txn, &restApiV1.ArtistFilter{SongId: &songEntity.SongId})
	if err != nil {
		return err
	}
	for ind, artist := range artists {
		if ind == 0 {
			artistNamesStr = artist.Name
		} else {
			artistNamesStr += ", " + artist.Name
		}
	}
	tag.SetArtist(artistNamesStr)

	// endregion

	// region Save tags
	if err = tag.Save(); err != nil {
		return err
	}
	// endregion

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return nil
}
