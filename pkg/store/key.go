package store

type Key string

func (k Key) Bytes() []byte {
	return []byte(k)
}

func (k Key) String() string {
	return string(k)
}
