package auth

import (
	"sync"

	"github.com/paynejacob/hotcereal/pkg/graph"
	"github.com/paynejacob/hotcereal/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

type TokenProvider struct {
	Store store.Store

	mu sync.RWMutex

	cache       map[string]*Token
	searchIndex *graph.Graph
	lookupToken map[string]*Token
}

func (p *TokenProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// initialize internal struct values
	p.cache = map[string]*Token{}
	p.searchIndex = graph.New()
	p.lookupToken = map[string]*Token{}

	// load values from store
	return p.Store.List(`auth::`, func(bytes []byte) error {
		o := Token{}

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		// write to the cache
		p.cache[o.Id] = &o

		// write to the search graph

		// add lookups
		p.lookupToken[o.Token] = &o

		return nil
	})
}

func (p *TokenProvider) GetKey(o *Token) string {
	// package::type::id
	return "auth::Token" + o.Id
}

func (p *TokenProvider) Get(k string) *Token {
	p.mu.RLock()

	if o, ok := p.cache[k]; ok {
		p.mu.RUnlock()
		return o
	}

	p.mu.RUnlock()
	return nil
}

func (p *TokenProvider) List() []*Token {
	rval := make([]*Token, 0)

	p.mu.RLock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	p.mu.RUnlock()
	return rval
}

func (p *TokenProvider) Search(query string) []*Token {
	results := make([]*Token, 0)

	p.mu.RLock()

	for _, key := range p.searchIndex.Search(query) {
		results = append(results, p.cache[key])
	}

	p.mu.RUnlock()
	return results
}

func (p *TokenProvider) GetByToken(v string) *Token {
	p.mu.RLock()

	if o, ok := p.lookupToken[v]; ok {
		p.mu.RUnlock()
		return o
	}

	p.mu.RUnlock()
	return nil
}

func (p *TokenProvider) Save(o *Token) error {
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

	// update lookups
	p.lookupToken[o.Token] = o

	p.mu.Unlock()

	return nil
}

func (p *TokenProvider) Delete(objs ...*Token) error {
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
		delete(p.lookupToken, obj.Token)

		delete(p.cache, obj.Id)
		p.searchIndex.Delete(obj.Id)
	}

	p.mu.Unlock()
	return nil
}
