package sound

import (
	"bytes"
	"errors"
	"fmt"
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
	duration += parts[0] * 60 * 60  // hours
	duration += parts[0] * 60 // minutes
	duration += parts[0] // seconds

	return duration, nil
}

func normalizeAudio(path string) error {
	return exec.Command("ffmpeg", fmt.Sprintf("ffmpeg -y -i %s -filter:a loudnorm -f mp3 %s", path, path)).Run()
}