package sound

import (
	"bytes"
	"encoding/gob"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/paynejacob/speakerbob/pkg/graph"
	"io"
	"sync"
	"time"
)

const (
	SoundKeyPrefix byte = iota
	AudioKeyPrefix
)

type Provider struct {
	maxSoundDuration time.Duration // the maximum length a sound can be

	mu          sync.RWMutex
	db          *badger.DB
	searchIndex *graph.Graph
}

func NewProvider(db *badger.DB, maxSoundDuration time.Duration) *Provider {
	return &Provider{
		mu:               sync.RWMutex{},
		maxSoundDuration: maxSoundDuration,
		db:               db,
		searchIndex:      graph.NewGraph(),
	}
}

// Sounds
func (p *Provider) SearchSounds(tokens [][]byte) (sounds []Sound, err error) {
	keys := make(map[string]bool, 0)

	if len(tokens) == 0 {
		tokens = append(tokens, []byte{})
	}

	p.mu.RLock()

	for _, token := range tokens {
		for _, b := range p.searchIndex.Search(token) {
			if b[0] == SoundKeyPrefix {
				keys[string(b[1:])] = true
			}
		}
	}

	p.mu.RUnlock()

	// we load the sounds separately to prevent loading them twice
	sounds = make([]Sound, len(keys))
	i := 0
	for key := range keys {
		sounds[i].Id = key
		if err = p.GetSound(&sounds[i]); err != nil {
			return
		}
		i++
	}

	return
}

func (p *Provider) GetSound(sound *Sound) error {
	var err error
	var item *badger.Item
	var buf bytes.Buffer

	err = p.db.View(func(txn *badger.Txn) error {
		item, err = txn.Get(sound.Key())
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			buf.Write(val)
			return nil
		})

		return err
	})

	if err != nil {
		return err
	}

	err = gob.NewDecoder(&buf).Decode(&sound)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) AllSounds() (sounds []Sound, err error) {
	var item *badger.Item
	var sound Sound

	prefix := []byte{SoundKeyPrefix}

	err = p.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item = it.Item()

			sound = Sound{}

			err = item.Value(func(val []byte) error {
				return gob.NewDecoder(bytes.NewReader(val)).Decode(&sound)
			})

			sounds = append(sounds, sound)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}

func (p *Provider) GetSoundAudio(sound Sound, w io.Writer) error {
	return p.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(sound.AudioKey())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			_, err = w.Write(val)
			return err
		})
	})
}

func (p *Provider) CreateSound(filename string, data io.ReadCloser) (sound Sound, err error) {
	var buf bytes.Buffer

	sound = NewSound()

	err = normalizeAudio(filename, p.maxSoundDuration, data, &buf)
	if err != nil {
		return
	}

	durationBuf := buf
	sound.Duration, err = getAudioDuration(&durationBuf)
	if err != nil {
		return
	}

	err = p.db.Update(func(txn *badger.Txn) error {
		_ = txn.Set(sound.Key(), sound.Bytes())
		_ = txn.Set(sound.AudioKey(), buf.Bytes())

		return nil
	})

	return
}

func (p *Provider) SaveSound(sound Sound) (err error) {
	var buf bytes.Buffer

	err = gob.NewEncoder(&buf).Encode(sound)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.db.Update(func(txn *badger.Txn) error {
		return txn.Set(sound.Key(), sound.Bytes())
	})

	if err != nil {
		return
	}

	for _, token := range graph.Tokenize(sound.Name) {
		p.searchIndex.Write(token, sound.Key())
	}

	return nil
}

func (p *Provider) DeleteSound(sound Sound) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.db.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(sound.Key())
		_ = txn.Delete(sound.AudioKey())

		return nil
	})

	if err != nil {
		return
	}

	p.searchIndex.Delete(sound.Key())

	return
}

func (p *Provider) HydrateSearch() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sounds, err := p.AllSounds()
	if err != nil {
		return err
	}

	for i := range sounds {
		if sounds[i].Name == "" {
			continue
		}

		for _, token := range graph.Tokenize(sounds[i].Name) {
			p.searchIndex.Write(token, sounds[i].Key())
		}
	}

	return nil
}
