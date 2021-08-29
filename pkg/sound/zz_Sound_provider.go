package sound

import (
	"sync"

	"github.com/paynejacob/hotcereal/pkg/graph"
	"github.com/paynejacob/hotcereal/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

type SoundProvider struct {
	Store store.Store

	mu sync.RWMutex

	cache       map[string]*Sound
	searchIndex *graph.Graph
}

func (p *SoundProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// initialize internal struct values
	p.cache = map[string]*Sound{}
	p.searchIndex = graph.New()

	// load values from store
	return p.Store.List(`sound::`, func(bytes []byte) error {
		o := Sound{}

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		// write to the cache
		p.cache[o.Id] = &o

		// write to the search graph
		p.searchIndex.Write(graph.Tokenize(o.Name), o.Id)

		// add lookups

		return nil
	})
}

func (p *SoundProvider) GetKey(o *Sound) string {
	// package::type::id
	return "sound::Sound" + o.Id
}

func (p *SoundProvider) Get(k string) *Sound {
	p.mu.RLock()

	if o, ok := p.cache[k]; ok {
		p.mu.RUnlock()
		return o
	}

	p.mu.RUnlock()
	return nil
}

func (p *SoundProvider) List() []*Sound {
	rval := make([]*Sound, 0)

	p.mu.RLock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	p.mu.RUnlock()
	return rval
}

func (p *SoundProvider) Search(query string) []*Sound {
	results := make([]*Sound, 0)

	p.mu.RLock()

	for _, key := range p.searchIndex.Search(query) {
		results = append(results, p.cache[key])
	}

	p.mu.RUnlock()
	return results
}

func (p *SoundProvider) Save(o *Sound) error {
	p.mu.Lock()

	// persist the object to the store
	body, err := msgpack.Marshal(o)
	if err = p.Store.Save(p.GetKey(o), body); err != nil {
		p.mu.Unlock()
		return err
	}

	// update the cache
	p.cache[o.Id] = o

	// update the search index
	p.searchIndex.Write(graph.Tokenize(o.Name), o.Id)

	// update lookups

	p.mu.Unlock()

	return nil
}

func (p *SoundProvider) Delete(objs ...*Sound) error {
	p.mu.Lock()

	keys := make([]string, len(objs))
	for i, obj := range objs {
		keys[i] = p.GetKey(obj)
	}

	// delete from the persistence layer
	if err := p.Store.Delete(keys...); err != nil {
		p.mu.Unlock()
		return err
	}

	for _, obj := range objs {
		// ensure the fields match the stored fields
		obj = p.Get(obj.Id)

		// cleanup lookups

		delete(p.cache, obj.Id)
		p.searchIndex.Delete(obj.Id)
	}

	p.mu.Unlock()
	return nil
}
