package sound

import (
	"fmt"

	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/spf13/cobra"
)

var deleteId string

var deleteCommand = &cobra.Command{
	Use:           "delete",
	Short:         "Delete a sound, cascading to any groups that reference it.",
	RunE:          runDelete,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	deleteCommand.Flags().StringVar(&deleteId, "id", "", "ID of the sound to delete.")
	_ = deleteCommand.MarkFlagRequired("id")
}

func runDelete(*cobra.Command, []string) error {
	soundProvider, groupProvider, closeStore, err := openProviders()
	if err != nil {
		return err
	}
	defer closeStore()

	return deleteSound(soundProvider, groupProvider, deleteId)
}

func deleteSound(soundProvider *psound.SoundProvider, groupProvider *psound.GroupProvider, id string) error {
	sound := soundProvider.Get(id)
	if sound == nil {
		return fmt.Errorf("sound %q not found", id)
	}

	if _, err := psound.DeleteSoundWithGroups(groupProvider, soundProvider, sound); err != nil {
		return fmt.Errorf("failed to delete sound %q: %w", id, err)
	}

	return nil
}
