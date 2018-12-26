package internal

import (
	"hash/fnv"
)

type Set map[uint64]string

func (s Set) Add(value string) {
	h := fnv.New64()
	_, _ = h.Write([]byte(value))

	s[h.Sum64()] = value
}

func (s Set) Remove(value string) {
	h := fnv.New64()
	_, _ = h.Write([]byte(value))

	s[h.Sum64()] = value
	delete(s, h.Sum64())
}

func (s Set) Union(b Set) Set {
	newSet := Set{}

	for hash, conn := range s {
		newSet[hash] = conn
	}

	for hash, conn := range b {
		newSet[hash] = conn
	}

	return newSet
}

func (s Set) Values() []string {
	values := make([]string, 0)

	for _, value := range s {
		values = append(values, value)
	}

	return values
}

func (s Set) Empty() bool {
	return len(s) > 0
}
