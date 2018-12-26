package api

import "speakerbob/internal"

type SearchResult interface {
	Type() string
	Key() string
	IndexValue() string
	Object() interface{}
}

type SearchBackend interface {
	UpdateResult(value SearchResult) error
	Remove(key string) error
	Search(query string, n int) ([]SearchResult, error)
}

type SearchMemoryBackend struct {
	values map[string]SearchResult
	index  map[string]internal.Set
}

func NewSearchMemoryBackend() *SearchMemoryBackend {
	return &SearchMemoryBackend{values: make(map[string]SearchResult, 0), index: make(map[string]internal.Set, 0)}
}

func (b SearchMemoryBackend) UpdateResult(value SearchResult) error {
	for i := 1; i < len(value.IndexValue()); i++ {
		subKey := value.IndexValue()[:i]

		if _, ok := b.index[subKey]; !ok {
			b.index[subKey] = internal.Set{}
		}

		b.index[subKey].Add(value.Key())
	}

	b.values[value.Key()] = value

	return nil
}

func (b SearchMemoryBackend) Remove(key string) error {
	for i := 0; i < len(key); i++ {
		subKey := key[:i]

		if _, ok := b.index[subKey]; ok {
			b.index[subKey].Remove(key)

			if b.index[subKey].Empty() {
				delete(b.index, subKey)
			}
		}
	}

	delete(b.values, key)

	return nil
}

func (b SearchMemoryBackend) Search(query string, n int) ([]SearchResult, error) {
	keys := make([]string, 0)
	results := make([]SearchResult, 0)

	if rs, ok := b.index[query]; ok {
		if len(rs) > n {
			keys = rs.Values()[:n]
		} else {
			keys = rs.Values()
		}

		for _, k := range keys {
			results = append(results, b.values[k])
		}
	}

	return results, nil
}
