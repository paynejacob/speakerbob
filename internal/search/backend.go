package search

type Result interface {
	Type() string
	Key() string
	IndexValue() string
	Object() interface{}
}

type Backend interface {
	UpdateResult(value Result) error
	Remove(key string) error
	Search(query string, n int) ([]Result, error)
}

type MemoryBackend struct {
	values map[string]Result
	index  map[string]Set
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{values: make(map[string]Result, 0), index: make(map[string]Set, 0)}
}

func (b MemoryBackend) UpdateResult(value Result) error {
	for i := 1; i < len(value.IndexValue()); i++ {
		subKey := value.IndexValue()[:i]

		if _, ok := b.index[subKey]; !ok {
			b.index[subKey] = Set{}
		}

		b.index[subKey].Add(value.Key())
	}

	b.values[value.Key()] = value

	return nil
}

func (b MemoryBackend) Remove(key string) error {
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

func (b MemoryBackend) Search(query string, n int) ([]Result, error) {
	keys := make([]string, 0)
	results := make([]Result, 0)

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
