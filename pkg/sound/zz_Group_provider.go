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
	return p.Store.List(p.TypeKey(), func(bytes []byte) error {
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

func (p *GroupProvider) Get(id string) *Group {
	p.mu.RLock()

	if o, ok := p.cache[id]; ok {
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

	for _, id := range p.searchIndex.Search(query) {
		results = append(results, p.cache[id])
	}

	p.mu.RUnlock()
	return results
}

func (p *GroupProvider) Save(o *Group) error {
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
	p.searchIndex.Write(graph.Tokenize(o.Name), o.Id)

	// update lookups

	p.mu.Unlock()

	return nil
}

func (p *GroupProvider) Delete(objs ...*Group) error {
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

func (p *GroupProvider) TypeKey() store.TypeKey {
	return store.TypeKey{
		Body:          "soundGroup",
		PackageLength: 5,
		TypeLength:    5,
	}
}

func (p *GroupProvider) ObjectKey(o *Group) store.ObjectKey {
	return store.ObjectKey{
		TypeKey:  p.TypeKey(),
		IdLength: len(o.Id),
	}
}

func (p *GroupProvider) FieldKey(o *Group, fieldName string) store.FieldKey {
	return store.FieldKey{
		ObjectKey:   p.ObjectKey(o),
		FieldLength: len(fieldName),
	}
}

var _ msgpack.CustomEncoder = (*Group)(nil)
var _ msgpack.CustomDecoder = (*Group)(nil)

func (s *Group) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(
		s.Id,
		s.CreatedAt,
		s.Name,
		s.Duration,
		s.SoundIds,
	)
}

func (s *Group) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(
		s.Id,
		s.CreatedAt,
		s.Name,
		s.Duration,
		s.SoundIds,
	)
}
