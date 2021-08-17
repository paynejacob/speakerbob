package auth

import (
	"fmt"
	"sync"

	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/paynejacob/speakerbob/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
)

// DO NOT EDIT THIS CODE IS GENERATED

const TokenProviderKeyPrefix = "auth.Token"

type TokenProvider struct {
	Store store.Store

	mu    sync.RWMutex
	cache map[string]*Token

	searchIndex *graph.Graph
	lookupToken map[string]*Token
}

func NewTokenProvider(s store.Store) *TokenProvider {
	return &TokenProvider{
		Store:       s,
		mu:          sync.RWMutex{},
		cache:       map[string]*Token{},
		lookupToken: map[string]*Token{},
		searchIndex: graph.NewGraph(),
	}
}

func (p *TokenProvider) Get(k string) *Token {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if o, ok := p.cache[k]; ok {
		return o
	}

	return &Token{}
}

func (p *TokenProvider) List() []*Token {
	rval := make([]*Token, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, o := range p.cache {
		rval = append(rval, o)
	}

	return rval
}

func (p *TokenProvider) Search(query string) []*Token {
	results := make([]*Token, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, keyBytes := range p.searchIndex.Search([]byte(query)) {
		results = append(results, p.cache[string(keyBytes)])
	}

	return results
}

func (p *TokenProvider) Save(o *Token) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := msgpack.Marshal(o)

	if err = p.Store.Save(p.GetKey(o), body); err != nil {
		return err
	}

	p.cache[o.Id] = o

	p.lookupToken[o.Token] = o

	return nil
}

func (p *TokenProvider) Delete(o *Token) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Store.Delete(p.GetKey(o)); err != nil {
		return err
	}

	o = p.Get(o.Id)

	delete(p.lookupToken, o.Token)

	delete(p.cache, o.Id)
	p.searchIndex.Delete([]byte(o.Id))

	return nil
}

func (p *TokenProvider) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Store.List([]byte(TokenProviderKeyPrefix), func(bytes []byte) error {
		o := Token{}

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		p.cache[o.Id] = &o

		p.lookupToken[o.Token] = &o

		return nil
	})
}

func (p *TokenProvider) GetByToken(v string) *Token {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if o, ok := p.lookupToken[v]; ok {
		return o
	}

	return &Token{}
}

func (p *TokenProvider) GetKey(o *Token) store.Key {
	return store.Key(fmt.Sprintf("%s:%s", TokenProviderKeyPrefix, o.Id))
}
