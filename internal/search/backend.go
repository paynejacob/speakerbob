package search

import (
	"hash/fnv"
)

type ResultSet map[uint64]string

func (s ResultSet) Add(key string, value string) {
	h := fnv.New64()
	_, _ = h.Write([]byte(key))

	s[h.Sum64()] = value
}

func (s ResultSet) Remove(key string) {
	h := fnv.New64()
	_, _ = h.Write([]byte(key))

	s[h.Sum64()] = key
	delete(s, h.Sum64())
}

func (s ResultSet) Union(b ResultSet) ResultSet {
	newSet := ResultSet{}

	for hash, conn := range s {
		newSet[hash] = conn
	}

	for hash, conn := range b {
		newSet[hash] = conn
	}

	return newSet
}

func (s ResultSet) Values() []string {
	values := make([]string, len(s))

	for _, value := range s {
		values = append(values, value)
	}

	return values
}

func (s ResultSet) Empty() bool {
	return len(s) > 0
}

type Backend interface {
	Update(key string, value string)
	Remove(key string)
	Search(query string, n int) []string
}

type MemoryBackend struct {
	index map[string]ResultSet
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{index: make(map[string]ResultSet)}
}

func (b MemoryBackend) Update(key string, value string) {
	for i := 0; i < len(key); i++ {
		subKey := key[:i]

		if _, ok := b.index[subKey]; !ok {
			b.index[subKey] = ResultSet{}
		}

		b.index[subKey].Add(key, value)
	}
}

func (b MemoryBackend) Remove(key string) {
	for i := 0; i < len(key); i++ {
		subKey := key[:i]

		if _, ok := b.index[subKey]; ok {
			b.index[subKey].Remove(key)

			if b.index[subKey].Empty() {
				delete(b.index, subKey)
			}
		}
	}
}

func (b MemoryBackend) Search(query string, n int) []string {
	if rs, ok := b.index[query]; ok {
		if len(rs) > n {
			return rs.Values()[:n]
		}
		return rs.Values()
	}

	return []string{}
}

