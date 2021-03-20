package sound

import (
	"fmt"
	"github.com/tcolgate/mp3"
	"io"
	"os/exec"
	"strings"
	"time"
)

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
