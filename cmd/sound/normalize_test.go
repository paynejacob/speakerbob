package sound

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSounds(t *testing.T) {
	provider := newTestSoundProvider(t)
	audioFile := writeTestAudioFile(t)

	s1, err := createSound(provider, "", audioFile, 10*time.Second)
	require.NoError(t, err)
	_, err = createSound(provider, "", audioFile, 10*time.Second)
	require.NoError(t, err)

	t.Run("single sound by id", func(t *testing.T) {
		require.NoError(t, normalizeSounds(provider, s1.Id, false, 10*time.Second))
	})

	t.Run("all sounds", func(t *testing.T) {
		require.NoError(t, normalizeSounds(provider, "", true, 10*time.Second))
	})

	t.Run("neither id nor all", func(t *testing.T) {
		assert.Error(t, normalizeSounds(provider, "", false, 10*time.Second))
	})

	t.Run("both id and all", func(t *testing.T) {
		assert.Error(t, normalizeSounds(provider, s1.Id, true, 10*time.Second))
	})

	t.Run("unknown id", func(t *testing.T) {
		assert.Error(t, normalizeSounds(provider, "does-not-exist", false, 10*time.Second))
	})
}
