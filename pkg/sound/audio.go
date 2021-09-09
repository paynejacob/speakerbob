package sound

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var specialCharacterRegexp = regexp.MustCompile(`[^a-zA-Z0-9\\s.? ]+`)

var durationRegexp = regexp.MustCompile(`time=(?P<h>\d+):(?P<m>\d+):(?P<s>\d+).(?P<ms>\d+)`)

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

	return duration, nil
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
