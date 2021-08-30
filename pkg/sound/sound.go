package sound

import (
	"bytes"
	"github.com/google/uuid"
	"io"
	"strings"
	"time"
)

//go:generate go run github.com/paynejacob/hotcereal providergen github.com/paynejacob/speakerbob/pkg/sound.Sound
type Sound struct {
	Id        string    `json:"id,omitempty" hotcereal:"key"`
	CreatedAt time.Time `json:"created_at,omitempty"`

	Name     string        `json:"name,omitempty" hotcereal:"searchable"`
	Duration time.Duration `json:"duration,omitempty"`
	Hidden   bool          `json:"-"`
	Audio 	 []byte `json:"-" hotcereal:"lazy"`
}

func NewSound() Sound {
	return Sound{
		Id:        strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt: time.Now(),
		Hidden:    true,
	}
}

func (p *SoundProvider) NewSound(filename string, audio io.ReadCloser, maxDuration time.Duration) (*Sound, error) {
	var err error
	var buf bytes.Buffer

	sound := NewSound()

	err = normalizeAudio(filename, maxDuration, audio, &buf)
	if err != nil {
		return nil, err
	}

	durationBuf := buf
	sound.Duration, err = getAudioDuration(&durationBuf)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.Save(&sound)
	if err != nil {
		return nil, err
	}

	err = p.WriteAudio(&sound, &buf)
	if err != nil {
		return nil, err
	}

	return &sound, err
}

func (p *SoundProvider) NewTTSSound(text string, maxDuration time.Duration) (*Sound, error) {
	var err error
	var buf bytes.Buffer
	var normBuf bytes.Buffer

	// create a new sound
	sound := NewSound()
	sound.Hidden = true

	// codegen audio
	err = tts(text, &buf)
	if err != nil {
		return nil, err
	}

	// normalize audio
	err = normalizeAudio("f.wav", maxDuration, &buf, &normBuf)
	if err != nil {
		return nil, err
	}

	// get the audio duration
	durationBuf := buf
	sound.Duration, err = getAudioDuration(&durationBuf)
	if err != nil {
		return nil, err
	}

	err = p.Save(&sound)
	if err != nil {
		return nil, err
	}

	err = p.WriteAudio(&sound, &buf)
	if err != nil {
		return nil, err
	}

	return &sound, err
}
