package sound

import (
	"errors"
	"github.com/minio/minio-go"
	"io"
	"net/http"
	"os"
	"path"
)

type Backend interface {
	ServeRedirect() bool
	PutSound(sound Sound, file io.Reader) error
	GetSound(sound Sound) (io.Reader, error)
	RedirectSound(sound Sound, w http.ResponseWriter, r *http.Request) error
	RemoveSound(sound Sound) error
}

type LocalBackend struct {
	Directory string
}

func NewlocalBackend(directory string) *LocalBackend {
	return &LocalBackend{directory}
}

func (b LocalBackend) ServeRedirect() bool {
	return false
}

func (b LocalBackend) PutSound(sound Sound, file io.Reader) error {
	newFile, err := os.Create(path.Join(b.Directory, sound.Id))
	if err != nil {
		return err
	}

	if _, err = io.Copy(newFile, file); err != nil {
		return err
	}

	return nil
}

func (b LocalBackend) GetSound(sound Sound) (io.Reader, error) {
	if file, err := os.Open(path.Join(b.Directory, sound.Id)); err != nil {
		return nil, err
	} else {
		return file, nil
	}
}

func (b LocalBackend) RedirectSound(sound Sound, w http.ResponseWriter, r *http.Request) error {
	return errors.New("not implemented")
}

func (b LocalBackend) RemoveSound(sound Sound) error {
	return os.Remove(path.Join(b.Directory, sound.Id))
}

type MinioBackend struct {
	bucketName string

	minioClient *minio.Client
}

func NewMinioBackend(url string, accessID string, accessKey string, useSSL bool, bucketName string) *MinioBackend {
	client, err := minio.New(url, accessID, accessKey, useSSL)
	if err != nil {
		panic("failed to configure minio")
	}
	ensureBucket(bucketName, client)
	return &MinioBackend{bucketName, client}
}

func (b MinioBackend) ServeRedirect() bool {
	return true
}

func (b MinioBackend) PutSound(sound Sound, file io.Reader) error {
	if _, err := b.minioClient.PutObject(b.bucketName, sound.Id, file, -1, minio.PutObjectOptions{}); err != nil {
		return err
	}

	return nil
}

func (b MinioBackend) GetSound(sound Sound) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (b MinioBackend) RedirectSound(sound Sound, w http.ResponseWriter, r *http.Request) error {
	if redirURL, err := b.minioClient.PresignedGetObject(b.bucketName, sound.Id, 60, nil); err != nil {
		return err
	} else {
		http.Redirect(w, r, redirURL.String(), http.StatusTemporaryRedirect)
	}

	return nil
}

func (b MinioBackend) RemoveSound(sound Sound) error {
	return b.minioClient.RemoveObject(b.bucketName, sound.Id)
}
