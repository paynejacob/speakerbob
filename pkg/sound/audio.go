package sound

import (
	"bytes"
	"fmt"
	"github.com/tcolgate/mp3"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var specialCharacterRegexp = regexp.MustCompile(`[^a-zA-Z0-9\\s.? ]+`)

var durationRegexp = regexp.MustCompile(`Duration: (?P<h>\d+):(?P<m>\d+):(?P<s>\d+).(?P<ms>\d+)`)

func normalizeAudio(filename string, maxDuration time.Duration, r io.Reader, w io.Writer) (time.Duration, error) {
	var output bytes.Buffer

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
	cmd.Stderr = &output
	cmd.Stdin = r

	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	var duration time.Duration
	match := durationRegexp.FindSubmatch(output.Bytes())
	if len(match) > 0 {
		for i, name := range durationRegexp.SubexpNames() {
			subD, _ := time.ParseDuration(string(match[i]) + name)

			duration += subD

			if duration > maxDuration {
				duration = maxDuration
				break
			}
		}
	}

	return duration, nil
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

func tts(text string, w io.Writer) error {
	cmd := exec.Command(
		"flite",
		"-voice", "slt",
		"-t", specialCharacterRegexp.ReplaceAllString(text, ""),
		"-o", "/dev/stdout")
	cmd.Stdout = w

	return cmd.Run()
}
