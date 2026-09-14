package sound

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

var durationRegexp = regexp.MustCompile(`time=(?P<h>\d+):(?P<m>\d+):(?P<s>\d+).(?P<ms>\d+)`)

func normalizeAudio(filename string, maxDuration time.Duration, r io.Reader, w io.Writer) (time.Duration, error) {
	var output bytes.Buffer

	// Containers such as MP4/M4A store their index (the moov atom) anywhere in
	// the file and require a seekable input to demux, which a pipe can't
	// provide. Writing the upload to a temp file first fixes that, and also
	// lets ffmpeg auto-probe the real format instead of trusting a format
	// guessed from the client-supplied filename (which was previously derived
	// with strings.Split(filename, ".")[1] - wrong for any filename containing
	// more than one "." before its extension, e.g. "foo.mp3cut.net.mp3").
	tmp, err := os.CreateTemp("", "speakerbob-upload-*")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, r); err != nil {
		return 0, err
	}

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-hide_banner",
		"-loglevel", "info",
		"-i", tmp.Name(),
		"-ss", "0",
		"-t", fmt.Sprintf("%.0f", maxDuration.Seconds()),
		"-c:a", "libmp3lame",
		"-filter:a", "loudnorm",
		"-f", "mp3",
		"pipe:1")
	cmd.Stdout = w
	cmd.Stderr = &output

	err = cmd.Run()
	if err != nil {
		return 0, err
	}

	var duration time.Duration
	matches := durationRegexp.FindAllSubmatch(output.Bytes(), -1)
	if matches != nil {
		match := matches[len(matches)-1]
		for i, name := range durationRegexp.SubexpNames() {
			if i == 0 {
				continue
			}

			subD, _ := time.ParseDuration(string(match[i]) + name)

			duration += subD

			if duration > maxDuration {
				duration = maxDuration
				break
			}
		}
	}

	if duration == 0 {
		logrus.Debugf("failed to get duration for %s, output: %s", filename, output.String())
	}

	return duration, nil
}
