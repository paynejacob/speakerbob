package sound

import (
	"io"
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
	return p.Store.List(p.TypeKey(), func(bytes []byte) error {
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

func (p *SoundProvider) Get(id string) *Sound {
	p.mu.RLock()

	if o, ok := p.cache[id]; ok {
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

	for _, id := range p.searchIndex.Search(query) {
		results = append(results, p.cache[id])
	}

	p.mu.RUnlock()
	return results
}

func (p *SoundProvider) Save(o *Sound) error {
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

func (p *SoundProvider) Delete(objs ...*Sound) error {
	p.mu.Lock()

	var keys []store.Key

	for _, obj := range objs {
		keys = append(keys,
			p.ObjectKey(obj),
			p.FieldKey(obj, "Audio"),
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

func (p *SoundProvider) ReadAudio(o *Sound, w io.Writer) error {
	return p.Store.ReadLazy(p.FieldKey(o, "Audio"), w)
}

func (p *SoundProvider) WriteAudio(o *Sound, r io.Reader) error {
	return p.Store.WriteLazy(p.FieldKey(o, "Audio"), r)
}

func (p *SoundProvider) TypeKey() store.TypeKey {
	return store.TypeKey{
		Body:          "soundSound",
		PackageLength: 5,
		TypeLength:    5,
	}
}

func (p *SoundProvider) ObjectKey(o *Sound) store.ObjectKey {
	return store.ObjectKey{
		TypeKey:  p.TypeKey(),
		IdLength: len(o.Id),
	}
}

func (p *SoundProvider) FieldKey(o *Sound, fieldName string) store.FieldKey {
	return store.FieldKey{
		ObjectKey:   p.ObjectKey(o),
		FieldLength: len(fieldName),
	}
}

var _ msgpack.CustomEncoder = (*Sound)(nil)
var _ msgpack.CustomDecoder = (*Sound)(nil)

func (s *Sound) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(
		s.Id,
		s.CreatedAt,
		s.Name,
		s.Duration,
		s.Hidden,
	)
}

func (s *Sound) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(
		s.Id,
		s.CreatedAt,
		s.Name,
		s.Duration,
		s.Hidden,
	)
}
