package sound

import (
	"testing"

	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSound(t *testing.T) {
	soundProvider := newTestSoundProvider(t)
	groupProvider := newTestGroupProvider(t)

	sound := psound.NewSound()
	require.NoError(t, soundProvider.Save(&sound))

	group := psound.NewGroup()
	group.SoundIds = []string{sound.Id}
	require.NoError(t, groupProvider.Save(&group))

	t.Run("cascades to referencing groups", func(t *testing.T) {
		require.NoError(t, deleteSound(soundProvider, groupProvider, sound.Id))

		assert.Nil(t, soundProvider.Get(sound.Id))
		assert.Nil(t, groupProvider.Get(group.Id))
	})

	t.Run("unknown id", func(t *testing.T) {
		assert.Error(t, deleteSound(soundProvider, groupProvider, "does-not-exist"))
	})
}
