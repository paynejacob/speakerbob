package sound

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/tags"
	"github.com/tcolgate/mp3"
	"io"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const downloadUrlTTL = 60 * time.Second

type Store struct {
	*minio.Client
	bucketName string

	maxSoundDuration       time.Duration // the maximum length a sound can be
	soundCreateGracePeriod time.Duration // how long before an unnamed sound is deleted

	m      sync.RWMutex
	sounds map[string]Sound
}

func NewStore(client *minio.Client, bucketName string, maxSoundDuration time.Duration) *Store {
	return &Store{
		Client:                 client,
		bucketName:             bucketName,
		m:                      sync.RWMutex{},
		maxSoundDuration:       maxSoundDuration,
		soundCreateGracePeriod: 1 * time.Hour,
		sounds:                 map[string]Sound{}}
}

func (p *Store) Create(filename string, data io.ReadCloser) (sound Sound, err error) {
	var buf bytes.Buffer

	sound = NewSound()

	err = normalizeAudio(filename, p.maxSoundDuration, data, &buf)
	if err != nil {
		return
	}

	durationBuf := buf
	sound.Duration, err = getAudioDuration(&durationBuf)
	if err != nil {
		return
	}

	p.m.Lock()
	defer p.m.Unlock()

	uploadBuf := buf
	_, err = p.PutObject(context.TODO(), p.bucketName, sound.Id, &uploadBuf, int64(uploadBuf.Len()), minio.PutObjectOptions{ContentType: "audio/mp3"})
	if err != nil {
		return
	}

	p.sounds[sound.Id] = sound

	return sound, nil
}

func (p *Store) Get(soundId string) (Sound, error) {
	p.m.RLock()
	defer p.m.RUnlock()

	if sound, ok := p.sounds[soundId]; ok {
		return sound, nil
	}

	var sound Sound

	p.m.RUnlock()
	p.m.Lock()
	defer p.m.Unlock()

	sound.Id = soundId
	if err := p.loadFromS3(&sound); err != nil {
		return sound, err
	}

	return sound, nil
}

func (p *Store) Save(sound Sound) (err error) {
	p.m.Lock()
	defer p.m.Unlock()

	err = p.saveToS3(&sound)
	if err != nil {
		return
	}

	p.sounds[sound.Id] = sound

	return
}

func (p *Store) Delete(sound Sound) (err error) {
	p.m.Lock()
	defer p.m.Unlock()

	err = p.deleteFromS3(sound)
	if err != nil {
		return
	}

	delete(p.sounds, sound.Id)

	return
}

func (p *Store) GetDownloadUrl(sound Sound) (*url.URL, error) {
	return p.PresignedGetObject(context.TODO(), p.bucketName, sound.Id, downloadUrlTTL, url.Values{})
}

func (p *Store) All() []Sound {
	p.m.RLock()
	defer p.m.RUnlock()

	sounds := make([]Sound, 0)

	for _, sound := range p.sounds {
		if sound.Name == "" {
			continue
		}

		sounds = append(sounds, sound)
	}

	return sounds
}

func (p *Store) UninitializedSounds() []Sound {
	p.m.RLock()
	defer p.m.RUnlock()

	sounds := make([]Sound, 0)

	for _, sound := range p.sounds {
		if sound.Name != "" {
			continue
		}

		sounds = append(sounds, sound)
	}

	return sounds
}

func (p *Store) InitializeCache() error {
	p.m.Lock()
	defer p.m.Unlock()

	var _tags *tags.Tags
	var err error

	for obj := range p.ListObjects(context.TODO(), p.bucketName, minio.ListObjectsOptions{}) {
		if obj.Err != nil {
			return obj.Err
		}

		_tags, err = p.GetObjectTagging(context.TODO(), p.bucketName, obj.Key, minio.GetObjectTaggingOptions{})
		if err != nil {
			return err
		}

		sound := Sound{Id: obj.Key}

		soundFromTagMap(&sound, _tags.ToMap())

		p.sounds[obj.Key] = sound
	}

	return nil
}

// Storage Actions
func (p *Store) saveToS3(sound *Sound) (err error) {
	soundTagMap := make(map[string]string, 0)

	soundToTagMap(sound, soundTagMap)

	soundTags, err := tags.MapToObjectTags(soundTagMap)
	if err != nil {
		return err
	}

	return p.PutObjectTagging(context.TODO(), p.bucketName, sound.Id, soundTags, minio.PutObjectTaggingOptions{})
}

func (p *Store) loadFromS3(sound *Sound) error {
	tagging, err := p.GetObjectTagging(context.TODO(), p.bucketName, sound.Id, minio.GetObjectTaggingOptions{})
	if err != nil {
		return err
	}

	soundFromTagMap(sound, tagging.ToMap())

	return nil
}

func (p *Store) deleteFromS3(sound Sound) error {
	return p.RemoveObject(context.TODO(), p.bucketName, sound.Id, minio.RemoveObjectOptions{})
}

// Serialization
func soundFromTagMap(sound *Sound, tagMap map[string]string) {
	if val, ok := tagMap["speakerbob.com/Sound/CreatedAt"]; ok {
		sound.CreatedAt, _ = time.Parse(time.RFC3339, val)
	}

	if val, ok := tagMap["speakerbob.com/Sound/Name"]; ok {
		sound.Name = val
	}

	if val, ok := tagMap["speakerbob.com/Sound/Duration"]; ok {
		sound.Duration, _ = time.ParseDuration(val)
	}

	if val, ok := tagMap["speakerbob.com/Sound/NSFW"]; ok {
		sound.NSFW = val == "true"
	}
}

func soundToTagMap(s *Sound, o map[string]string) {
	o["speakerbob.com/Sound/Id"] = s.Id
	o["speakerbob.com/Sound/Name"] = s.Name
	o["speakerbob.com/Sound/CreatedAt"] = s.CreatedAt.Format(time.RFC3339)
	o["speakerbob.com/Sound/Duration"] = s.Duration.String()
	o["speakerbob.com/Sound/NSFW"] = "false"
	if s.NSFW {
		o["speakerbob.com/Sound/NSFW"] = "true"
	}
}

// Audio Normalization
func normalizeAudio(filename string, maxDuration time.Duration, r io.ReadCloser, w io.Writer) error {
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-hide_banner",
		"-loglevel", "info",
		"-f", strings.Split(filename, ".")[1],
		"-i", "pipe:0",
		"-ss", "0",
		"-t", fmt.Sprintf("%.0f", maxDuration.Seconds()),
		"-c:a", "libmp3lame",
		"-filter:a", "loudnorm",
		"-f", "mp3",
		"pipe:1")
	cmd.Stdout = w
	cmd.Stdin = r

	return cmd.Run()
}

func getAudioDuration(r io.Reader) (time.Duration, error) {
	var t int64
	var f mp3.Frame
	var skipped int

	d := mp3.NewDecoder(r)

	for {
		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		t = t + f.Duration().Milliseconds()
	}

	return time.Duration(t) * time.Millisecond, nil
}
