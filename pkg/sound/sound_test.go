package sound

import (
	"bytes"
	"testing"
	"time"

	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimal valid empty-data WAV header, same fixture used in service_test.go's TestCreateSound
var testWavBytes = []byte{82, 73, 70, 70, 36, 0, 0, 0, 87, 65, 86, 69, 102, 109, 116, 32, 16, 0, 0, 0, 1, 0, 1, 0, 68, 172, 0, 0, 136, 88, 1, 0, 2, 0, 16, 0, 100, 97, 116, 97, 0, 0, 0, 0}

func TestRenormalize(t *testing.T) {
	provider := &SoundProvider{Store: memory.New()}
	require.NoError(t, provider.Initialize())

	t.Run("re-normalizes existing audio", func(t *testing.T) {
		sound := NewSound()
		require.NoError(t, provider.Save(&sound))
		require.NoError(t, provider.WriteAudio(&sound, bytes.NewReader(testWavBytes)))

		err := provider.Renormalize(&sound, 10*time.Second)
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, provider.ReadAudio(&sound, &buf))
		assert.NotEmpty(t, buf.Bytes())

		saved := provider.Get(sound.Id)
		assert.Equal(t, sound.Duration, saved.Duration)
	})

	t.Run("errors when audio is not readable", func(t *testing.T) {
		sound := NewSound()
		require.NoError(t, provider.Save(&sound))

		err := provider.Renormalize(&sound, 10*time.Second)
		assert.Error(t, err)
	})
}
