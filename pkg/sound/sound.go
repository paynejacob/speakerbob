package sound

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/paynejacob/speakerbob/pkg/graph"
	"github.com/paynejacob/speakerbob/pkg/store"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"strings"
	"time"
)

//go:generate go run github.com/paynejacob/speakerbob/pkg/store/provider --type=Sound --package=sound --indexed_fields=Name
type Sound struct {
	Id        string        `json:"id,omitempty"`
	CreatedAt time.Time     `json:"created_at,omitempty"`
	Name      string        `json:"name,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	Hidden    bool          `json:"-"`
}

func NewSound() Sound {
	return Sound{
		Id:        strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt: time.Now(),
		Hidden:    true,
	}
}

func (s *Sound) AudioKey() store.Key {
	return store.Key(fmt.Sprintf("audio+%s", getSoundKey(s)))
}

func (p *SoundProvider) NewSound(filename string, audio io.ReadCloser, maxDuration time.Duration) (sound Sound, err error) {
	var buf bytes.Buffer
	var soundBuf []byte

	sound = NewSound()

	err = normalizeAudio(filename, maxDuration, audio, &buf)
	if err != nil {
		return
	}

	durationBuf := buf
	sound.Duration, err = getAudioDuration(&durationBuf)
	if err != nil {
		return
	}

	soundBuf, _ = msgpack.Marshal(sound)

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.Store.BulkSave(map[store.Key][]byte{
		store.Key(sound.Id): soundBuf,
		sound.AudioKey():    buf.Bytes(),
	})
	if err != nil {
		return
	}

	p.cache[sound.Id] = &sound
	p.index.Write(graph.Tokenize(sound.Name), []byte(sound.Id))

	return
}

func (p *SoundProvider) NewTTSSound(text string, maxDuration time.Duration) (*Sound, error) {
	var err error
	var buf bytes.Buffer
	var normBuf bytes.Buffer
	var soundBuf []byte

	// create a new sound
	sound := NewSound()
	sound.Hidden = true

	// codegen audio
	err = tts(text, &buf)
	if err != nil {
		return &sound, err
	}

	// normalize audio
	err = normalizeAudio("f.wav", maxDuration, &buf, &normBuf)
	if err != nil {
		return &sound, err
	}

	// get the audio duration
	durationBuf := buf
	sound.Duration, err = getAudioDuration(&durationBuf)
	if err != nil {
		return &sound, err
	}

	soundBuf, _ = msgpack.Marshal(&sound)

	// persist to db
	err = p.Store.BulkSave(map[store.Key][]byte{
		getSoundKey(&sound): soundBuf,
		sound.AudioKey():    normBuf.Bytes(),
	})

	return &sound, err
}

func (p *SoundProvider) GetAudio(sound *Sound, w io.Writer) (err error) {
	var b []byte

	b, err = p.Store.Get(sound.AudioKey())
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}
