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
	return p.Store.List(p.TypeKey(), func(bytes []byte) error {
		var o User

		if err := msgpack.Unmarshal(bytes, &o); err != nil {
			return err
		}

		// write to the cache
		p.cache[o.Id] = &o

		// write to the search graph

		// add lookups
		p.lookupEmail[o.Email] = &o
		for _, v := range o.Principals {
			p.lookupPrincipals[v] = &o
		}

		return nil
	})
}

func (p *UserProvider) Get(id string) *User {
	p.mu.RLock()

	if o, ok := p.cache[id]; ok {
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

	for _, id := range p.searchIndex.Search(query) {
		results = append(results, p.cache[id])
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
	if err = p.Store.Save(p.ObjectKey(o), body); err != nil {
		p.mu.Unlock()
		return err
	}

	// update the cache
	p.cache[o.Id] = o

	// update the search index

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

func (p *UserProvider) TypeKey() store.TypeKey {
	return store.TypeKey{
		Body:          "authUser",
		PackageLength: 4,
		TypeLength:    4,
	}
}

func (p *UserProvider) ObjectKey(o *User) store.ObjectKey {
	k := store.ObjectKey{
		TypeKey:  p.TypeKey(),
		IdLength: len(o.Id),
	}

	k.Body += o.Id
	return k
}

func (p *UserProvider) FieldKey(o *User, fieldName string) store.FieldKey {
	k := store.FieldKey{
		ObjectKey:   p.ObjectKey(o),
		FieldLength: len(fieldName),
	}

	k.Body += fieldName
	return k
}

var _ msgpack.CustomEncoder = (*User)(nil)
var _ msgpack.CustomDecoder = (*User)(nil)

func (s *User) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(
		s.Id,
		s.CreatedAt,
		s.Email,
		s.Principals,
		s.Preferences,
	)
}

func (s *User) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(
		&s.Id,
		&s.CreatedAt,
		&s.Email,
		&s.Principals,
		&s.Preferences,
	)
}
