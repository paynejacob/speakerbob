package api

import (
	"errors"
	"github.com/minio/minio-go"
	"io"
	"net/http"
	"os"
	"path"
)

type SoundBackend interface {
	ServeRedirect() bool
	PutSound(sound Sound, file io.Reader) error
	GetSound(sound Sound) (io.Reader, error)
	RedirectSound(sound Sound, w http.ResponseWriter, r *http.Request) error
	RemoveSound(sound Sound) error
}

type SoundLocalBackend struct {
	Directory string
}

func NewSoundLocalBackend(directory string) *SoundLocalBackend {
	_ = os.MkdirAll(directory, os.ModePerm)
	return &SoundLocalBackend{directory}
}

func (b SoundLocalBackend) ServeRedirect() bool {
	return false
}

func (b SoundLocalBackend) PutSound(sound Sound, file io.Reader) error {
	newFile, err := os.Create(path.Join(b.Directory, sound.Id))
	if err != nil {
		return err
	}

	if _, err = io.Copy(newFile, file); err != nil {
		return err
	}

	return nil
}

func (b SoundLocalBackend) GetSound(sound Sound) (io.Reader, error) {
	if file, err := os.Open(path.Join(b.Directory, sound.Id)); err != nil {
		return nil, err
	} else {
		return file, nil
	}
}

func (b SoundLocalBackend) RedirectSound(sound Sound, w http.ResponseWriter, r *http.Request) error {
	return errors.New("not implemented")
}

func (b SoundLocalBackend) RemoveSound(sound Sound) error {
	return os.Remove(path.Join(b.Directory, sound.Id))
}

type SoundMinioBackend struct {
	bucketName string

	minioClient *minio.Client
}

func NewSoundMinioBackend(url string, accessID string, accessKey string, useSSL bool, bucketName string) *SoundMinioBackend {
	client, err := minio.New(url, accessID, accessKey, useSSL)
	if err != nil {
		panic("failed to configure minio")
	}
	ensureBucket(bucketName, client)
	return &SoundMinioBackend{bucketName, client}
}

func (b SoundMinioBackend) ServeRedirect() bool {
	return true
}

func (b SoundMinioBackend) PutSound(sound Sound, file io.Reader) error {
	if _, err := b.minioClient.PutObject(b.bucketName, sound.Id, file, -1, minio.PutObjectOptions{}); err != nil {
		return err
	}

	return nil
}

func (b SoundMinioBackend) GetSound(sound Sound) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (b SoundMinioBackend) RedirectSound(sound Sound, w http.ResponseWriter, r *http.Request) error {
	if redirURL, err := b.minioClient.PresignedGetObject(b.bucketName, sound.Id, 60, nil); err != nil {
		return err
	} else {
		http.Redirect(w, r, redirURL.String(), http.StatusTemporaryRedirect)
	}

	return nil
}

func (b SoundMinioBackend) RemoveSound(sound Sound) error {
	return b.minioClient.RemoveObject(b.bucketName, sound.Id)
}
