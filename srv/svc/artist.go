package svc

import (
	"encoding/json"
	"github.com/dgraph-io/badger"
	"lyra/restApiV1"
	"lyra/tool"
	"strings"
	"time"
)

func (s *Service) ReadArtists(externalTrn *badger.Txn, filter *restApiV1.ArtistFilter) ([]*restApiV1.Artist, error) {
	artists := []*restApiV1.Artist{}

	opts := badger.DefaultIteratorOptions
	switch filter.Order {
	case restApiV1.ArtistOrderByArtistName:
		opts.Prefix = []byte(artistNameArtistIdPrefix)
		opts.PrefetchValues = false
	case restApiV1.ArtistOrderByUpdateTs:
		opts.Prefix = []byte(artistUpdateTsArtistIdPrefix)
		opts.PrefetchValues = false
	default:
		opts.Prefix = []byte(artistIdPrefix)
	}

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	if filter.Order == restApiV1.ArtistOrderByUpdateTs {
		it.Seek([]byte(artistUpdateTsArtistIdPrefix + indexTs(filter.FromTs)))
	} else {
		it.Rewind()
	}

	for ; it.Valid(); it.Next() {
		var artist *restApiV1.Artist

		switch filter.Order {
		case restApiV1.ArtistOrderByArtistName,
			restApiV1.ArtistOrderByUpdateTs:
			key := it.Item().KeyCopy(nil)

			artistId := strings.Split(string(key), ":")[2]
			var e error
			artist, e = s.ReadArtist(txn, artistId)
			if e != nil {
				return nil, e
			}
		default:
			encodedArtist, e := it.Item().ValueCopy(nil)
			if e != nil {
				return nil, e
			}
			e = json.Unmarshal(encodedArtist, &artist)
			if e != nil {
				return nil, e
			}
		}

		artists = append(artists, artist)

	}

	return artists, nil
}

func (s *Service) ReadArtist(externalTrn *badger.Txn, artistId string) (*restApiV1.Artist, error) {
	var artist *restApiV1.Artist

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	item, e := txn.Get(getArtistIdKey(artistId))
	if e != nil {
		if e == badger.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, e
	}
	encodedArtist, e := item.ValueCopy(nil)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal(encodedArtist, &artist)
	if e != nil {
		return nil, e
	}

	return artist, nil
}

func (s *Service) CreateArtist(externalTrn *badger.Txn, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
	var artist *restApiV1.Artist

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	// Create artist
	now := time.Now().UnixNano()

	artist = &restApiV1.Artist{
		Id:         tool.CreateUlid(),
		CreationTs: now,
		UpdateTs:   now,
		ArtistMeta: *artistMeta,
	}

	encodedArtist, _ := json.Marshal(artist)
	e := txn.Set(getArtistIdKey(artist.Id), encodedArtist)
	if e != nil {
		return nil, e
	}

	// Create artist name Index
	e = txn.Set(getArtistNameArtistIdKey(artist.Name, artist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Create updateTs Index
	e = txn.Set(getArtistUpdateTsArtistIdKey(artist.UpdateTs, artist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artist, nil
}

func (s *Service) UpdateArtist(externalTrn *badger.Txn, artistId string, artistMeta *restApiV1.ArtistMeta) (*restApiV1.Artist, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	artist, err := s.ReadArtist(txn, artistId)

	if err != nil {
		return nil, err
	}

	artistOldName := artist.Name
	artistOldUpdateTs := artist.UpdateTs
	artist.ArtistMeta = *artistMeta
	artist.UpdateTs = time.Now().UnixNano()

	// Update artist
	encodedArtist, _ := json.Marshal(artist)
	e := txn.Set(getArtistIdKey(artist.Id), encodedArtist)
	if e != nil {
		return nil, e
	}

	// Update artist name index
	e = txn.Delete(getArtistNameArtistIdKey(artistOldName, artist.Id))
	if e != nil {
		return nil, e
	}
	e = txn.Set(getArtistNameArtistIdKey(artist.Name, artist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Update updateTs Index
	e = txn.Delete(getArtistUpdateTsArtistIdKey(artistOldUpdateTs, artist.Id))
	if e != nil {
		return nil, e
	}
	e = txn.Set(getArtistUpdateTsArtistIdKey(artist.UpdateTs, artist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Update tags in songs content
	songIds, e := s.GetSongIdsFromArtistId(txn, artistId)
	if e != nil {
		return nil, e
	}

	for _, songId := range songIds {
		s.UpdateSong(txn, songId, nil, &artistId)
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artist, nil
}

func (s *Service) DeleteArtist(externalTrn *badger.Txn, artistId string) (*restApiV1.Artist, error) {
	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	deleteTs := time.Now().UnixNano()

	artist, e := s.ReadArtist(txn, artistId)

	if e != nil {
		return nil, e
	}

	// Check songs link
	songIds, e := s.GetSongIdsFromArtistId(txn, artistId)
	if e != nil {
		return nil, e
	}
	if len(songIds) > 0 {
		return nil, ErrDeleteArtistWithSongs
	}

	// Delete artist name index
	e = txn.Delete(getArtistNameArtistIdKey(artist.Name, artistId))
	if e != nil {
		return nil, e
	}

	// Delete artist updateTs index
	e = txn.Delete(getArtistUpdateTsArtistIdKey(artist.UpdateTs, artistId))
	if e != nil {
		return nil, e
	}

	// Delete artist
	e = txn.Delete(getArtistIdKey(artistId))
	if e != nil {
		return nil, e
	}

	// Archive artistId
	e = txn.Set(getArtistDeleteTsArtistIdKey(deleteTs, artist.Id), nil)
	if e != nil {
		return nil, e
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artist, nil
}

func (s *Service) GetDeletedArtistIds(externalTrn *badger.Txn, fromTs int64) ([]string, error) {

	artistIds := []string{}

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(artistDeleteTsArtistIdPrefix)
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Seek([]byte(artistDeleteTsArtistIdPrefix + indexTs(fromTs))); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		artistId := strings.Split(string(key), ":")[2]

		artistIds = append(artistIds, artistId)

	}

	return artistIds, nil
}

func (s *Service) GetArtistIdsByName(externalTrn *badger.Txn, artistName string) ([]string, error) {

	var artistIds []string

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(artistNameArtistIdPrefix + indexString(artistName) + ":")
	opts.PrefetchValues = false

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(false)
		defer txn.Discard()
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {

		key := it.Item().KeyCopy(nil)

		artistId := strings.Split(string(key), ":")[2]

		artistIds = append(artistIds, artistId)

	}

	return artistIds, nil
}

func (s *Service) getArtistIdsFromArtistNames(externalTrn *badger.Txn, artistNames []string) ([]string, error) {

	var artistIds []string

	// Check available transaction
	txn := externalTrn
	if txn == nil {
		txn = s.Db.NewTransaction(true)
		defer txn.Discard()
	}

	for _, artistName := range artistNames {
		artistName = normalizeString(artistName)
		if artistName != "" {
			artistIdsFromName, err := s.GetArtistIdsByName(txn, artistName)
			if err != nil {
				return nil, err
			}
			var artistId string
			if len(artistIdsFromName) > 0 {
				// Link the song to an existing artist
				artistId = artistIdsFromName[0]
			} else {
				// Create the artist before linking it to the song
				artist, err := s.CreateArtist(txn, &restApiV1.ArtistMeta{Name: artistName})
				if err != nil {
					return nil, err
				}
				artistId = artist.Id
			}
			artistIds = append(artistIds, artistId)
		}
	}

	// Commit transaction
	if externalTrn == nil {
		txn.Commit()
	}

	return artistIds, nil
}

func isArtistIdsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
