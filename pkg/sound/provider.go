package sound

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"github.com/paynejacob/speakerbob/pkg/graph"
	"io"
	"sync"
	"time"
)

const (
	SoundKeyPrefix byte = iota
	AudioKeyPrefix
	GroupKeyPrefix
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

func (p *Provider) Search(tokens [][]byte) (sounds []Sound, groups []Group, err error) {
	var sound Sound
	var group Group
	keys := make(map[string]bool, 0)
	sounds = make([]Sound, 0)
	groups = make([]Group, 0)

	if len(tokens) == 0 {
		tokens = append(tokens, []byte{})
	}

	p.mu.RLock()

	for _, token := range tokens {
		for _, b := range p.searchIndex.Search(token) {
			keys[string(b)] = true
		}
	}

	p.mu.RUnlock()

	for key := range keys {
		if key[0] == SoundKeyPrefix {
			sound.Id = key[1:]
			if err = p.GetSound(&sound); err != nil {
				return
			}
			sounds = append(sounds, sound)
		} else if key[0] == GroupKeyPrefix {
			group.Id = key[1:]
			if err = p.GetGroup(&group); err != nil {
				return
			}
			groups = append(groups, group)
		}
	}

	return
}

// Sounds
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

func (p *Provider) SaveSound(sound Sound) (err error) {
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
	var item *badger.Item
	prefix := []byte{GroupKeyPrefix}
	groupKeys := make([][]byte, 0)

	// find groups that use the sound
	err = p.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item = it.Item()

			var group Group

			err = item.Value(func(val []byte) error {
				return gob.NewDecoder(bytes.NewReader(val)).Decode(&group)
			})
			if err != nil {
				return err
			}

			for i := range group.SoundIds {
				if group.SoundIds[i] == sound.Id {
					groupKeys = append(groupKeys, item.Key())
				}
			}
		}

		return nil
	})

	err = p.db.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(sound.Key())
		_ = txn.Delete(sound.AudioKey())
		for i := range groupKeys {
			_ = txn.Delete(groupKeys[i])
		}

		return nil
	})

	if err != nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.searchIndex.Delete(sound.Key())
	for i := range groupKeys {
		p.searchIndex.Delete(groupKeys[i])
	}

	return
}

func (p *Provider) HydrateSearch() error {
	var item *badger.Item
	var sound Sound
	var group Group
	var err error

	p.mu.Lock()
	defer p.mu.Unlock()

	return p.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek([]byte{}); it.ValidForPrefix([]byte{}); it.Next() {
			item = it.Item()

			switch item.Key()[0] {
			case SoundKeyPrefix:
				err = item.Value(func(val []byte) error {
					return gob.NewDecoder(bytes.NewReader(val)).Decode(&sound)
				})

				if err != nil {
					return err
				}

				if sound.Name == "" {
					continue
				}

				for _, token := range graph.Tokenize(sound.Name) {
					p.searchIndex.Write(token, sound.Key())
				}
				break
			case GroupKeyPrefix:
				err = item.Value(func(val []byte) error {
					return gob.NewDecoder(bytes.NewReader(val)).Decode(&group)
				})

				if err != nil {
					return err
				}

				for _, token := range graph.Tokenize(group.Name) {
					p.searchIndex.Write(token, group.Key())
				}
				break
			}
		}

		return nil
	})
}

// Groups
func (p *Provider) CreateGroup(name string, sounds []Sound) (Group, error) {
	var group Group

	group = NewGroup()

	group.Name = name
	group.SoundIds = make([]string, len(sounds))

	for i := range sounds {
		group.NSFW = sounds[i].NSFW || group.NSFW
		group.Duration += sounds[i].Duration

		group.SoundIds[i] = sounds[i].Id
	}

	err := p.db.Update(func(txn *badger.Txn) error {
		return txn.Set(group.Key(), group.Bytes())
	})

	if err != nil {
		return group, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for _, token := range graph.Tokenize(group.Name) {
		p.searchIndex.Write(token, group.Key())
	}

	return group, nil
}

func (p *Provider) GetGroup(group *Group) error {
	var err error
	var item *badger.Item
	var buf bytes.Buffer

	err = p.db.View(func(txn *badger.Txn) error {
		item, err = txn.Get(group.Key())
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

	err = gob.NewDecoder(&buf).Decode(&group)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) SaveGroup(group Group) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.db.Update(func(txn *badger.Txn) error {
		return txn.Set(group.Key(), group.Bytes())
	})

	if err != nil {
		return
	}

	for _, token := range graph.Tokenize(group.Name) {
		p.searchIndex.Write(token, group.Key())
	}

	return nil
}

func (p *Provider) DeleteGroup(group Group) (err error) {
	err = p.db.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(group.Key())

		return nil
	})

	if err != nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.searchIndex.Delete(group.Key())

	return
}
