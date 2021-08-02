package sound

import (
	"fmt"
	"github.com/tcolgate/mp3"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var specialCharacterRegexp = regexp.MustCompile(`[^a-zA-Z0-9\\s.? ]+`)

func normalizeAudio(filename string, maxDuration time.Duration, r io.Reader, w io.Writer) error {
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

func tts(text string, w io.Writer) error {
	cmd := exec.Command(
		"flite",
		"-voice", "slt",
		"-t", specialCharacterRegexp.ReplaceAllString(text, ""),
		"-o", "/dev/stdout")
	cmd.Stdout = w

	return cmd.Run()
}
