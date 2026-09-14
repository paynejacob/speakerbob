package sound

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/stretchr/testify/require"
)

func newTestSoundProvider(t *testing.T) *psound.SoundProvider {
	t.Helper()

	p := &psound.SoundProvider{Store: memory.New()}
	require.NoError(t, p.Initialize())

	return p
}

func newTestGroupProvider(t *testing.T) *psound.GroupProvider {
	t.Helper()

	p := &psound.GroupProvider{Store: memory.New()}
	require.NoError(t, p.Initialize())

	return p
}

// buildTestWav synthesizes a short mono PCM WAV containing an actual tone.
// A WAV with an empty data chunk (as used elsewhere in this repo's tests for
// single-pass create-only assertions) round-trips through ffmpeg once, but
// produces an empty/undecodable MP3 that a second normalization pass (as
// exercised by the CLI's normalize command) cannot read back in.
func buildTestWav(seconds float64, sampleRate int) []byte {
	numSamples := int(seconds * float64(sampleRate))
	data := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ {
		v := int16(3000 * math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate)))
		binary.LittleEndian.PutUint16(data[i*2:], uint16(v))
	}

	var buf bytes.Buffer
	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(36+len(data)))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // mono
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)

	return buf.Bytes()
}

func writeTestAudioFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.wav")
	require.NoError(t, os.WriteFile(path, buildTestWav(0.2, 8000), 0o600))

	return path
}
