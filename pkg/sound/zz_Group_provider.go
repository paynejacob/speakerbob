package sound

import (
	"sync"

	"github.com/paynejacob/hotcereal/pkg/graph"
	"github.com/paynejacob/hotcereal/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

type GroupProvider struct {
	Store store.Store

	mu sync.RWMutex

	cache       map[string]*Group
	searchIndex *graph.Graph
}

func (p *GroupProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// initialize internal struct values
	p.cache = map[string]*Group{}
	p.searchIndex = graph.New()

	// load values from store
	return p.Store.List(`sound::`, func(bytes []byte) error {
		o := Group{}

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

func (p *GroupProvider) GetKey(o *Group) string {
	// package::type::id
	return "sound::Group" + o.Id
}

func (p *GroupProvider) Get(k string) *Group {
	p.mu.RLock()

	if o, ok := p.cache[k]; ok {
		p.mu.RUnlock()
		return o
	}

	p.mu.RUnlock()
	return nil
}

func (p *GroupProvider) List() []*Group {
	rval := make([]*Group, 0)

	p.mu.RLock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	p.mu.RUnlock()
	return rval
}

func (p *GroupProvider) Search(query string) []*Group {
	results := make([]*Group, 0)

	p.mu.RLock()

	for _, key := range p.searchIndex.Search(query) {
		results = append(results, p.cache[key])
	}

	p.mu.RUnlock()
	return results
}

func (p *GroupProvider) Save(o *Group) error {
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

func (p *GroupProvider) Delete(objs ...*Group) error {
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
