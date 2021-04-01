package sound

import (
	"bytes"
	"encoding/gob"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Sound struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	NSFW     bool          `json:"nsfw"`
	Hidden   bool          `json:"-"`
}

func NewSound() Sound {
	return Sound{
		Id:        strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt: time.Now(),
		Hidden:    true,
	}
}

func (s Sound) Bytes() []byte {
	var buf bytes.Buffer

	_ = gob.NewEncoder(&buf).Encode(s)

	return buf.Bytes()
}

func (s Sound) Key() []byte {
	return append([]byte{SoundKeyPrefix}, []byte(s.Id)...)
}

func (s Sound) AudioKey() []byte {
	return append([]byte{AudioKeyPrefix}, []byte(s.Id)...)
}

type Group struct {
	Sound

	SoundIds []string `json:"sounds"`
}

func NewGroup() Group {
	var g Group

	g.Id = strings.Replace(uuid.New().String(), "-", "", 4)
	g.CreatedAt = time.Now()

	return g
}

func (g Group) Key() []byte {
	return append([]byte{GroupKeyPrefix}, []byte(g.Id)...)
}

func (g Group) Bytes() []byte {
	var buf bytes.Buffer

	_ = gob.NewEncoder(&buf).Encode(g)

	return buf.Bytes()
}
