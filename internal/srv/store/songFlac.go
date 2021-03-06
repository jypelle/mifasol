package store

import (
	"bytes"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
	"github.com/jmoiron/sqlx"
	"github.com/jypelle/mifasol/internal/srv/entity"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type MifasolMetaDataBlockVorbisComment struct {
	flacvorbis.MetaDataBlockVorbisComment
}

func (s *Store) createSongNewFromFlacContent(externalTrn *sqlx.Tx, content []byte, lastAlbumId restApiV1.AlbumId) (*restApiV1.SongNew, error) {

	// Extract song meta from tags
	flacFile, err := flac.ParseMetadata(bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}

	var cmt *flacvorbis.MetaDataBlockVorbisComment
	for _, meta := range flacFile.Meta {
		if meta.Type == flac.VorbisComment {
			cmt, err = flacvorbis.ParseFromMetaDataBlock(*meta)
			if err != nil {
				panic(err)
			}
		}
	}
	if cmt == nil {
		cmt = flacvorbis.New()
	}

	var songNew *restApiV1.SongNew

	var bitDepth = restApiV1.SongBitDepthUnknown
	var title = ""
	var albumId restApiV1.AlbumId = restApiV1.UnknownAlbumId
	var trackNumber *int64 = nil
	var publicationYear *int64 = nil
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

	// Extract bit depth
	streamInfoBlock, err := flacFile.GetStreamInfo()
	if err != nil {
		return nil, err
	}
	switch streamInfoBlock.BitDepth {
	case 16:
		bitDepth = restApiV1.SongBitDepth16
	case 24:
		bitDepth = restApiV1.SongBitDepth24
	default:
		bitDepth = restApiV1.SongBitDepthUnknown
	}

	// Extract title
	titles, err := cmt.Get(flacvorbis.FIELD_TITLE)
	if err != nil {
		return nil, err
	}
	if len(titles) > 0 {
		title = titles[0]
	}
	logrus.Debugf("Title: %s", title)

	// Extract album name
	albumName := ""
	albumNames, err := cmt.Get(flacvorbis.FIELD_ALBUM)
	if err != nil {
		return nil, err
	}
	if len(albumNames) > 0 {
		albumName = normalizeString(albumNames[0])
	}

	// Find Album Id
	albumId, err = s.getAlbumIdFromAlbumName(txn, albumName, lastAlbumId)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Album: %s", albumName)

	// Extract track number
	if albumId != restApiV1.UnknownAlbumId {
		trackNumbers, err := cmt.Get(flacvorbis.FIELD_TRACKNUMBER)
		if err != nil {
			return nil, err
		}

		if len(trackNumbers) > 0 {
			parsedTrackNumber, _ := strconv.ParseInt(normalizeString(trackNumbers[0]), 10, 64)
			if parsedTrackNumber > 0 {
				trackNumber = &parsedTrackNumber
			}
		}

	}

	if trackNumber != nil {
		logrus.Debugf("Track number: %d", *trackNumber)
	}

	// Extract year
	yearNumbers, err := cmt.Get(flacvorbis.FIELD_DATE)
	if err != nil {
		return nil, err
	}

	if len(yearNumbers) > 0 {
		parsedYearNumber, _ := strconv.ParseInt(normalizeString(yearNumbers[0]), 10, 64)
		if parsedYearNumber > 0 {
			publicationYear = &parsedYearNumber
		}
	}

	if publicationYear != nil {
		logrus.Debugf("Publication year: %d", *publicationYear)
	}

	// Extract artists
	vorbisArtistNames, err := cmt.Get(flacvorbis.FIELD_ARTIST)
	if err != nil {
		return nil, err
	}

	// Extract artists
	var artistNames []string
	for _, vorbisArtistName := range vorbisArtistNames {
		artistNames = append(artistNames, strings.FieldsFunc(vorbisArtistName, func(r rune) bool { return r == ',' || r == ';' || r == ';' })...)
	}

	// Find Artist IDs
	logrus.Debugf("Find artist ids")
	artistIds, err = s.getArtistIdsFromArtistNames(txn, artistNames)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Artists: %v", artistNames)

	songNew = &restApiV1.SongNew{
		SongMeta: restApiV1.SongMeta{
			Name:            title,
			Format:          restApiV1.SongFormatFlac,
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

func (s *Store) updateSongContentFlacTag(externalTrn *sqlx.Tx, songEntity *entity.SongEntity) error {

	// region Extract tags
	flacFile, err := flac.ParseFile(s.getSongFileName(songEntity.SongId, songEntity.Format))
	if err != nil {
		return err
	}

	var cmt *flacvorbis.MetaDataBlockVorbisComment
	var oldCmtKey = -1

	for key, meta := range flacFile.Meta {
		if meta.Type == flac.VorbisComment {
			cmt, err = flacvorbis.ParseFromMetaDataBlock(*meta)
			if err != nil {
				return err
			}
			oldCmtKey = key
		}
	}

	if cmt == nil {
		cmt = flacvorbis.New()
	}

	// endregion

	// region Update tags with song meta

	// Set title
	vorbisClean(cmt, flacvorbis.FIELD_TITLE)
	cmt.Add(flacvorbis.FIELD_TITLE, songEntity.Name)

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
	vorbisClean(cmt, flacvorbis.FIELD_ALBUM)
	vorbisClean(cmt, flacvorbis.FIELD_TRACKNUMBER)
	if songEntity.AlbumId != restApiV1.UnknownAlbumId {
		album, err := s.ReadAlbum(txn, songEntity.AlbumId)
		if err != nil {
			return err
		}
		cmt.Add(flacvorbis.FIELD_ALBUM, album.Name)

		if songEntity.TrackNumber.Valid {
			cmt.Add(flacvorbis.FIELD_TRACKNUMBER, strconv.FormatInt(songEntity.TrackNumber.Int64, 10))
		}
	}

	// Set publication date
	vorbisClean(cmt, flacvorbis.FIELD_DATE)
	if songEntity.PublicationYear.Valid {
		cmt.Add(flacvorbis.FIELD_DATE, strconv.FormatInt(songEntity.PublicationYear.Int64, 10))
	}

	// Set artists
	vorbisClean(cmt, flacvorbis.FIELD_ARTIST)
	artists, err := s.ReadArtists(txn, &restApiV1.ArtistFilter{SongId: &songEntity.SongId})
	if err != nil {
		return err
	}
	for _, artist := range artists {
		cmt.Add(flacvorbis.FIELD_ARTIST, artist.Name)
	}

	// endregion

	// region Save tags

	metaDataBlock := cmt.Marshal()
	if oldCmtKey != -1 {
		flacFile.Meta[oldCmtKey] = &metaDataBlock
	} else {
		flacFile.Meta = append(flacFile.Meta, &metaDataBlock)
	}
	flacFile.Save(s.getSongFileName(songEntity.SongId, songEntity.Format))

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	// endregion
	return nil
}

// vorbisClean remove all comments with field name specified by the key parameter
func vorbisClean(c *flacvorbis.MetaDataBlockVorbisComment, key string) error {
	res := make([]string, 0)
	for _, cmt := range c.Comments {
		p := strings.SplitN(cmt, "=", 2)
		if len(p) != 2 {
			return flacvorbis.ErrorMalformedComment
		}
		if !strings.EqualFold(p[0], key) {
			res = append(res, cmt)
		}
	}
	c.Comments = res

	return nil
}
