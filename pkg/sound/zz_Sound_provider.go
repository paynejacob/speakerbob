package sound

import (
	"fmt"
	"sync"

	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/paynejacob/speakerbob/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

const SoundProviderKeyPrefix = "sound.Sound"

type SoundProvider struct {
	Store store.Store

	mu    sync.RWMutex
	cache map[string]*Sound

	searchIndex *graph.Graph
}

func NewSoundProvider(s store.Store) *SoundProvider {
	return &SoundProvider{
		Store:       s,
		mu:          sync.RWMutex{},
		cache:       map[string]*Sound{},
		searchIndex: graph.NewGraph(),
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

	for _, keyBytes := range p.searchIndex.Search([]byte(query)) {
		results = append(results, p.cache[string(keyBytes)])
	}

	return results
}

func (p *SoundProvider) Save(o *Sound) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := msgpack.Marshal(o)

	if err = p.Store.Save(p.GetKey(o), body); err != nil {
		return err
	}

	p.cache[o.Id] = o

	p.searchIndex.Write(graph.Tokenize(o.Name), []byte(o.Id))

	return nil
}

func (p *SoundProvider) Delete(o *Sound) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Store.Delete(p.GetKey(o)); err != nil {
		return err
	}

	o = p.Get(o.Id)

	delete(p.cache, o.Id)
	p.searchIndex.Delete([]byte(o.Id))

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

		p.searchIndex.Write(graph.Tokenize(o.Name), []byte(o.Id))

		return nil
	})
}

func (p *SoundProvider) GetKey(o *Sound) store.Key {
	return store.Key(fmt.Sprintf("%s:%s", SoundProviderKeyPrefix, o.Id))
}
