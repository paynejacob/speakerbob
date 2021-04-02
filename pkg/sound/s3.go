package sound

import (
	"bytes"
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

// S3 Storage Actions
func loadFromS3(client *minio.Client, bucketName string, sound *Sound) error {
	tagging, err := client.GetObjectTagging(context.TODO(), bucketName, sound.Id, minio.GetObjectTaggingOptions{})
	if err != nil {
		return err
	}

	soundFromTagMap(sound, tagging.ToMap())

	return nil
}

func SyncFromS3(db *badger.DB, client *minio.Client, bucketName string) error {
	var err error
	var sound Sound
	var buf bytes.Buffer
	var dl *minio.Object

	for obj := range client.ListObjects(context.TODO(), bucketName, minio.ListObjectsOptions{}) {
		sound.Id = obj.Key

		if obj.Err != nil {
			return obj.Err
		}

		// Download the audio file
		dl, err = client.GetObject(context.TODO(), bucketName, sound.Id, minio.GetObjectOptions{})
		if err != nil {
			return err
		}

		// Copy the audio file
		_, err = io.Copy(&buf, dl)
		if err != nil {
			return err
		}

		if dl.Close() != nil {
			return err
		}

		// Get the sound data
		err = loadFromS3(client, bucketName, &sound)
		if err != nil {
			return err
		}

		// write to database
		err = db.Update(func(txn *badger.Txn) error {
			_, exists := txn.Get(sound.Key())
			if exists == badger.ErrKeyNotFound {
				logrus.Infof("importing: %s", sound.Name)
				_ = txn.Set(sound.Key(), sound.Bytes())
				_ = txn.Set(sound.AudioKey(), buf.Bytes())
			} else {
				logrus.Debugf("skipping: %s", sound.Name)
			}

			return nil
		})

		if err != nil {
			return err
		}

		buf.Reset()
	}

	return nil
}

// S3 Serialization
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
}
