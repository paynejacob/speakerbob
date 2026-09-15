package sound

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/paynejacob/speakerbob/pkg/service"
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
	Audio    []byte        `json:"-" hotcereal:"lazy"`
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

	sound.Duration, err = normalizeAudio(filename, maxDuration, audio, &buf)
	if err != nil {
		return nil, service.NotAcceptableError{SpeakerbobError: "unable to interpret audio format"}
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

// Renormalize re-runs audio normalization against a sound's already-stored
// audio, for cases where the normalization parameters change after upload
// (e.g. a duration limit change) rather than the audio itself. Unlike
// NewSound/NewTTSSound, audio is written before Duration is saved: this
// mutates an already-live record, so a write failure must not leave a
// previously-correct Duration pointing at now-stale audio.
func (p *SoundProvider) Renormalize(sound *Sound, maxDuration time.Duration) error {
	var audioBuf bytes.Buffer
	var normBuf bytes.Buffer
	var err error

	if err = p.ReadAudio(sound, &audioBuf); err != nil {
		return fmt.Errorf("failed to read audio for sound %q: %w", sound.Id, err)
	}

	sound.Duration, err = normalizeAudio(sound.Id, maxDuration, &audioBuf, &normBuf)
	if err != nil {
		return fmt.Errorf("failed to normalize audio for sound %q: %w", sound.Id, err)
	}

	if err = p.WriteAudio(sound, &normBuf); err != nil {
		return fmt.Errorf("failed to write audio for sound %q: %w", sound.Id, err)
	}

	if err = p.Save(sound); err != nil {
		return fmt.Errorf("failed to save sound %q: %w", sound.Id, err)
	}

	return nil
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
	sound.Duration, err = normalizeAudio("f.wav", maxDuration, &buf, &normBuf)
	if err != nil {
		return nil, err
	}

	err = p.Save(&sound)
	if err != nil {
		return nil, err
	}

	err = p.WriteAudio(&sound, &normBuf)
	if err != nil {
		return nil, err
	}

	return &sound, err
}
