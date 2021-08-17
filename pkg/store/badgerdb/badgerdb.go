package badgerdb

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/paynejacob/speakerbob/pkg/store"
)

type Store struct {
	DB *badger.DB
}

func (b Store) Get(key store.Key) ([]byte, error) {
	var err error
	var item *badger.Item
	var rval []byte

	err = b.DB.View(func(txn *badger.Txn) error {
		item, err = txn.Get(key.Bytes())
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			rval = val
			return nil
		})

		return nil
	})

	return rval, err
}

func (b Store) List(prefix []byte, process func([]byte) error) error {
	var item *badger.Item

	return b.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item = it.Item()

			if err := item.Value(process); err != nil {
				return err
			}
		}

		return nil
	})
}

func (b Store) Save(key store.Key, bytes []byte) error {
	return b.BulkSave(map[store.Key][]byte{key: bytes})
}

func (b Store) BulkSave(m map[store.Key][]byte) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		for k, v := range m {
			_ = txn.Set(k.Bytes(), v)
		}

		return nil
	})
}

func (b Store) Delete(keys ...store.Key) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		for i := range keys {
			_ = txn.Delete(keys[i].Bytes())
		}

		return nil
	})
}

func (b Store) Close() error {
	return b.DB.Close()
}
