package sound

import (
	"fmt"
	"sync"

	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/paynejacob/speakerbob/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

const GroupProviderKeyPrefix = "sound.Group"

type GroupProvider struct {
	Store store.Store

	mu    sync.RWMutex
	cache map[string]*Group

	searchIndex *graph.Graph
}

func NewGroupProvider(s store.Store) *GroupProvider {
	return &GroupProvider{
		Store:       s,
		mu:          sync.RWMutex{},
		cache:       map[string]*Group{},
		searchIndex: graph.NewGraph(),
	}
}

func (p *GroupProvider) Get(k string) *Group {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if o, ok := p.cache[k]; ok {
		return o
	}

	return &Group{}
}

func (p *GroupProvider) List() []*Group {
	rval := make([]*Group, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	return rval
}

func (p *GroupProvider) Search(query string) []*Group {
	results := make([]*Group, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, keyBytes := range p.searchIndex.Search([]byte(query)) {
		results = append(results, p.cache[string(keyBytes)])
	}

	return results
}

func (p *GroupProvider) Save(o *Group) error {
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

func (p *GroupProvider) Delete(o *Group) error {
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

func (p *GroupProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Store.List([]byte(GroupProviderKeyPrefix), func(bytes []byte) error {
		o := Group{}

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		p.cache[o.Id] = &o

		p.searchIndex.Write(graph.Tokenize(o.Name), []byte(o.Id))

		return nil
	})
}

func (p *GroupProvider) GetKey(o *Group) store.Key {
	return store.Key(fmt.Sprintf("%s:%s", GroupProviderKeyPrefix, o.Id))
}
