package store

type Store interface {
	Get(key Key) ([]byte, error)
	List(prefix []byte, process func([]byte) error) error

	Save(Key, []byte) error
	BulkSave(map[Key][]byte) error

	Delete(key ...Key) error

	Close() error
}
