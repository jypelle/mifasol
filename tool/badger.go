package tool

import (
	"github.com/dgraph-io/badger"
)

func ReplaceKey(txn *badger.Txn, old []byte, new []byte) error {

	if old != nil {
		e := txn.Delete(old)
		if e != nil {
			return e
		}
	}

	e := txn.Set(new, nil)
	if e != nil {
		return e
	}

	return nil
}
