package sound

import (
	"io"
	"os/exec"
	"regexp"
)

var specialCharacterRegexp = regexp.MustCompile(`[^a-zA-Z0-9\\s.? ]+`)

// TTSEngine synthesizes speech audio from text. flite is the only
// implementation today; the interface exists so an alternative engine
// (e.g. https://github.com/coqui-ai/TTS) can be added later without
// touching the rest of pkg/sound.
type TTSEngine interface {
	Synthesize(text string, w io.Writer) error
}

type fliteTTSEngine struct{}

func (fliteTTSEngine) Synthesize(text string, w io.Writer) error {
	cmd := exec.Command(
		"flite",
		"-voice", "slt",
		"-t", specialCharacterRegexp.ReplaceAllString(text, ""),
		"-o", "/dev/stdout")
	cmd.Stdout = w

	return cmd.Run()
}

var ttsEngine TTSEngine = fliteTTSEngine{}
