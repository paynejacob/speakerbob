package sound

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSound(t *testing.T) {
	provider := newTestSoundProvider(t)
	audioFile := writeTestAudioFile(t)

	t.Run("without a name stays hidden", func(t *testing.T) {
		sound, err := createSound(provider, "", audioFile, 10*time.Second)
		require.NoError(t, err)
		assert.True(t, sound.Hidden)
		assert.Empty(t, sound.Name)
	})

	t.Run("with a name is unhidden and named", func(t *testing.T) {
		sound, err := createSound(provider, "test-sound", audioFile, 10*time.Second)
		require.NoError(t, err)
		assert.False(t, sound.Hidden)
		assert.Equal(t, "test-sound", sound.Name)
	})

	t.Run("missing audio file", func(t *testing.T) {
		_, err := createSound(provider, "", "/nonexistent/path.wav", 10*time.Second)
		assert.Error(t, err)
	})
}
