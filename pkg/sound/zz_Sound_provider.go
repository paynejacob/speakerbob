package sound

import (
	"fmt"
	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/paynejacob/speakerbob/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
	"sync"
)

const SoundProviderKeyPrefix = "sound.Sound"

type SoundProvider struct {
	Store store.Store

	mu    sync.RWMutex
	cache map[string]*Sound
	index *graph.Graph
}

func NewSoundProvider(s store.Store) *SoundProvider {
	return &SoundProvider{
		Store: s,
		mu:    sync.RWMutex{},
		cache: map[string]*Sound{},
		index: graph.NewGraph(),
	}
}

func (p *SoundProvider) Get(k string) *Sound {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if o, ok := p.cache[k]; ok {
		return o
	}

	return &Sound{}
}

func (p *SoundProvider) List() []*Sound {
	rval := make([]*Sound, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	return rval
}

func (p *SoundProvider) Search(query string) []*Sound {
	results := make([]*Sound, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, keyBytes := range p.index.Search([]byte(query)) {
		results = append(results, p.cache[string(keyBytes)])
	}

	return results
}

func (p *SoundProvider) Save(o *Sound) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := msgpack.Marshal(o)

	if err = p.Store.Save(getSoundKey(o), body); err != nil {
		return err
	}

	p.cache[o.Id] = o

	p.index.Write(graph.Tokenize(o.Name), []byte(o.Id))

	return nil
}

func (p *SoundProvider) Delete(o *Sound) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Store.Delete(getSoundKey(o)); err != nil {
		return err
	}

	delete(p.cache, o.Id)
	p.index.Delete([]byte(o.Id))

	return nil
}

func (p *SoundProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Store.List([]byte(SoundProviderKeyPrefix), func(bytes []byte) error {
		o := Sound{}

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		p.cache[o.Id] = &o

		p.index.Write(graph.Tokenize(o.Name), []byte(o.Id))

		return nil
	})
}

func getSoundKey(o *Sound) store.Key {
	return store.Key(fmt.Sprintf("%s:%s", SoundProviderKeyPrefix, o.Id))
}
