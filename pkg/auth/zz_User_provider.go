package auth

import (
	"fmt"
	"sync"

	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/paynejacob/speakerbob/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

const UserProviderKeyPrefix = "auth.User"

type UserProvider struct {
	Store store.Store

	mu    sync.RWMutex
	cache map[string]*User

	searchIndex      *graph.Graph
	lookupEmail      map[string]*User
	lookupPrincipals map[Principal]*User
}

func NewUserProvider(s store.Store) *UserProvider {
	return &UserProvider{
		Store:            s,
		mu:               sync.RWMutex{},
		cache:            map[string]*User{},
		lookupEmail:      map[string]*User{},
		lookupPrincipals: map[Principal]*User{},
		searchIndex:      graph.NewGraph(),
	}
}

func (p *UserProvider) Get(k string) *User {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if o, ok := p.cache[k]; ok {
		return o
	}

	return &User{}
}

func (p *UserProvider) List() []*User {
	rval := make([]*User, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	return rval
}

func (p *UserProvider) Search(query string) []*User {
	results := make([]*User, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, keyBytes := range p.searchIndex.Search([]byte(query)) {
		results = append(results, p.cache[string(keyBytes)])
	}

	return results
}

func (p *UserProvider) Save(o *User) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := msgpack.Marshal(o)

	if err = p.Store.Save(p.GetKey(o), body); err != nil {
		return err
	}

	p.cache[o.Id] = o

	p.searchIndex.Write(graph.Tokenize(o.Name), []byte(o.Id))

	p.lookupEmail[o.Email] = o

	for _, v := range o.Principals {
		p.lookupPrincipals[v] = o
	}

	return nil
}

func (p *UserProvider) Delete(o *User) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Store.Delete(p.GetKey(o)); err != nil {
		return err
	}

	o = p.Get(o.Id)

	delete(p.lookupEmail, o.Email)

	for _, v := range o.Principals {
		delete(p.lookupPrincipals, v)
	}

	delete(p.cache, o.Id)
	p.searchIndex.Delete([]byte(o.Id))

	return nil
}

func (p *UserProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Store.List([]byte(UserProviderKeyPrefix), func(bytes []byte) error {
		o := User{}

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		p.cache[o.Id] = &o

		p.searchIndex.Write(graph.Tokenize(o.Name), []byte(o.Id))

		p.lookupEmail[o.Email] = &o

		for _, v := range o.Principals {
			p.lookupPrincipals[v] = &o
		}

		return nil
	})
}

func (p *UserProvider) GetByEmail(v string) *User {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if o, ok := p.lookupEmail[v]; ok {
		return o
	}

	return &User{}
}

func (p *UserProvider) GetByPrincipals(v Principal) *User {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if o, ok := p.lookupPrincipals[v]; ok {
		return o
	}

	return &User{}
}

func (p *UserProvider) GetKey(o *User) store.Key {
	return store.Key(fmt.Sprintf("%s:%s", UserProviderKeyPrefix, o.Id))
}
