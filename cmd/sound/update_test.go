package sound

import (
	"testing"

	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateSoundName(t *testing.T) {
	provider := newTestSoundProvider(t)

	sound := psound.NewSound()
	require.NoError(t, provider.Save(&sound))

	t.Run("valid name unhides the sound", func(t *testing.T) {
		require.NoError(t, updateSoundName(provider, sound.Id, "renamed"))

		updated := provider.Get(sound.Id)
		assert.Equal(t, "renamed", updated.Name)
		assert.False(t, updated.Hidden)
	})

	t.Run("empty name is rejected", func(t *testing.T) {
		assert.Error(t, updateSoundName(provider, sound.Id, ""))
	})

	t.Run("unknown id", func(t *testing.T) {
		assert.Error(t, updateSoundName(provider, "does-not-exist", "name"))
	})
}
