package auth

import (
	"sync"

	"github.com/paynejacob/hotcereal/pkg/graph"
	"github.com/paynejacob/hotcereal/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

type UserProvider struct {
	Store store.Store

	mu sync.RWMutex

	cache            map[string]*User
	searchIndex      *graph.Graph
	lookupEmail      map[string]*User
	lookupPrincipals map[Principal]*User
}

func (p *UserProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// initialize internal struct values
	p.cache = map[string]*User{}
	p.searchIndex = graph.New()
	p.lookupEmail = map[string]*User{}
	p.lookupPrincipals = map[Principal]*User{}

	// load values from store
	return p.Store.List(`auth::`, func(bytes []byte) error {
		o := User{}

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		// write to the cache
		p.cache[o.Id] = &o

		// write to the search graph
		p.searchIndex.Write(graph.Tokenize(o.Name), o.Id)

		// add lookups
		p.lookupEmail[o.Email] = &o
		for _, v := range o.Principals {
			p.lookupPrincipals[v] = &o
		}

		return nil
	})
}

func (p *UserProvider) GetKey(o *User) string {
	// package::type::id
	return "auth::User" + o.Id
}

func (p *UserProvider) Get(k string) *User {
	p.mu.RLock()

	if o, ok := p.cache[k]; ok {
		p.mu.RUnlock()
		return o
	}

	p.mu.RUnlock()
	return nil
}

func (p *UserProvider) List() []*User {
	rval := make([]*User, 0)

	p.mu.RLock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	p.mu.RUnlock()
	return rval
}

func (p *UserProvider) Search(query string) []*User {
	results := make([]*User, 0)

	p.mu.RLock()

	for _, key := range p.searchIndex.Search(query) {
		results = append(results, p.cache[key])
	}

	p.mu.RUnlock()
	return results
}

func (p *UserProvider) GetByEmail(v string) *User {
	p.mu.RLock()

	if o, ok := p.lookupEmail[v]; ok {
		p.mu.RUnlock()
		return o
	}

	p.mu.RUnlock()
	return nil
}
func (p *UserProvider) GetByPrincipals(v Principal) *User {
	p.mu.RLock()

	if o, ok := p.lookupPrincipals[v]; ok {
		p.mu.RUnlock()
		return o
	}

	p.mu.RUnlock()
	return nil
}

func (p *UserProvider) Save(o *User) error {
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
	p.lookupEmail[o.Email] = o
	for _, v := range o.Principals {
		p.lookupPrincipals[v] = o
	}

	p.mu.Unlock()

	return nil
}

func (p *UserProvider) Delete(objs ...*User) error {
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
		delete(p.lookupEmail, obj.Email)
		for _, v := range obj.Principals {
			delete(p.lookupPrincipals, v)
		}

		delete(p.cache, obj.Id)
		p.searchIndex.Delete(obj.Id)
	}

	p.mu.Unlock()
	return nil
}
