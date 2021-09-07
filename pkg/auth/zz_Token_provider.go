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
	return p.Store.List(p.TypeKey(), func(bytes []byte) error {
		var o Token

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

func (p *TokenProvider) Get(id string) *Token {
	p.mu.RLock()

	if o, ok := p.cache[id]; ok {
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

	for _, id := range p.searchIndex.Search(query) {
		results = append(results, p.cache[id])
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
	if err = p.Store.Save(p.ObjectKey(o), body); err != nil {
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

	var keys []store.Key

	for _, obj := range objs {
		keys = append(keys,
			p.ObjectKey(obj),
		)
	}

	// delete from the persistence layer
	if err := p.Store.Delete(keys...); err != nil {
		p.mu.Unlock()
		return err
	}

	var exists bool
	for _, obj := range objs {
		// ensure the fields match the stored fields
		obj, exists = p.cache[obj.Id]
		if !exists {
			continue
		}

		// cleanup lookups
		delete(p.lookupToken, obj.Token)

		delete(p.cache, obj.Id)
		p.searchIndex.Delete(obj.Id)
	}

	p.mu.Unlock()
	return nil
}

func (p *TokenProvider) TypeKey() store.TypeKey {
	return store.TypeKey{
		Body:          "authToken",
		PackageLength: 4,
		TypeLength:    5,
	}
}

func (p *TokenProvider) ObjectKey(o *Token) store.ObjectKey {
	k := store.ObjectKey{
		TypeKey:  p.TypeKey(),
		IdLength: len(o.Id),
	}

	k.Body += o.Id
	return k
}

func (p *TokenProvider) FieldKey(o *Token, fieldName string) store.FieldKey {
	k := store.FieldKey{
		ObjectKey:   p.ObjectKey(o),
		FieldLength: len(fieldName),
	}

	k.Body += fieldName
	return k
}

var _ msgpack.CustomEncoder = (*Token)(nil)
var _ msgpack.CustomDecoder = (*Token)(nil)

func (s *Token) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(
		s.Id,
		s.CreatedAt,
		s.Name,
		s.Token,
		s.Type,
		s.UserId,
		s.ExpiresAt,
	)
}

func (s *Token) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(
		&s.Id,
		&s.CreatedAt,
		&s.Name,
		&s.Token,
		&s.Type,
		&s.UserId,
		&s.ExpiresAt,
	)
}
