package sound

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTTSEngine struct {
	output []byte
	err    error
}

func (f fakeTTSEngine) Synthesize(_ string, w io.Writer) error {
	if f.err != nil {
		return f.err
	}
	_, err := w.Write(f.output)
	return err
}

func withTTSEngine(t *testing.T, engine TTSEngine) {
	t.Helper()
	original := ttsEngine
	ttsEngine = engine
	t.Cleanup(func() { ttsEngine = original })
}

func TestNewTTSSoundUsesConfiguredEngine(t *testing.T) {
	withTTSEngine(t, fakeTTSEngine{output: testWavBytes})

	provider := &SoundProvider{Store: memory.New()}
	require.NoError(t, provider.Initialize())

	sound, err := provider.NewTTSSound("hello world", 10*time.Second)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, provider.ReadAudio(sound, &buf))
	assert.NotEmpty(t, buf.Bytes())
	assert.True(t, sound.Hidden)
}

func TestNewTTSSoundPropagatesEngineError(t *testing.T) {
	withTTSEngine(t, fakeTTSEngine{err: errors.New("synthesis failed")})

	provider := &SoundProvider{Store: memory.New()}
	require.NoError(t, provider.Initialize())

	_, err := provider.NewTTSSound("hello world", 0)
	assert.ErrorContains(t, err, "synthesis failed")
}

func TestFliteTTSEngineSynthesize(t *testing.T) {
	if _, err := exec.LookPath("flite"); err != nil {
		t.Skip("flite not found on PATH")
	}

	var buf bytes.Buffer
	err := fliteTTSEngine{}.Synthesize("hello", &buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())
}
