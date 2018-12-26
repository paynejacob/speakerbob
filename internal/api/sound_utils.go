package api

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/minio/minio-go"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var DurationExpression = regexp.MustCompile("Duration: (\\d+:\\d+:\\d+)")

func getAudioDuration(path string) (int, error) {
	var out bytes.Buffer
	var parts [3]int
	var duration int

	cmd := exec.Command("ffmpeg", "-i", path)
	cmd.Stderr = &out
	_ = cmd.Run()

	res := DurationExpression.FindSubmatch(out.Bytes())
	if len(res) != 2 {
		return -1, errors.New("failed to parse ffmpeg output")
	}

	// hh:mm:ss.d
	for i, part := range strings.Split(string(res[1]), ":") {
		val, err := strconv.Atoi(part)
		if err != nil {
			return -1, err
		}

		parts[i] = val
	}
	duration += parts[0] * 60 * 60 // hours
	duration += parts[1] * 60      // minutes
	duration += parts[2]           // seconds

	return duration, nil
}

func normalizeAudio(path string) (string, error) {
	normalPath := fmt.Sprintf("%s.normal", path)
	return normalPath, exec.Command("ffmpeg", "-y", "-i", path, "-filter:a", "loudnorm", "-f", "wav", normalPath).Run()
}

func ensureBucket(soundBucketName string, minio *minio.Client) {
	err := minio.MakeBucket(soundBucketName, "us-east-1")
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := minio.BucketExists(soundBucketName)
		if err != nil || !exists {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Audio bucket was created %s\n", soundBucketName)
	}
}

func hashSpeakName(text string) string {
	h := sha256.New()

	_, _ = h.Write([]byte(text))

	return string(h.Sum(nil))
}
