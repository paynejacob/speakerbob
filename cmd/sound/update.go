package sound

import (
	"fmt"

	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/spf13/cobra"
)

var (
	updateId   string
	updateName string
)

var updateCommand = &cobra.Command{
	Use:           "update",
	Short:         "Update a sound's name.",
	RunE:          runUpdate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	updateCommand.Flags().StringVar(&updateId, "id", "", "ID of the sound to update.")
	updateCommand.Flags().StringVar(&updateName, "name", "", "New name for the sound (1-49 characters).")
	_ = updateCommand.MarkFlagRequired("id")
	_ = updateCommand.MarkFlagRequired("name")
}

func runUpdate(*cobra.Command, []string) error {
	soundProvider, _, closeStore, err := openProviders()
	if err != nil {
		return err
	}
	defer closeStore()

	return updateSoundName(soundProvider, updateId, updateName)
}

func updateSoundName(soundProvider *psound.SoundProvider, id, name string) error {
	sound := soundProvider.Get(id)
	if sound == nil {
		return fmt.Errorf("sound %q not found", id)
	}

	return setSoundName(soundProvider, sound, name)
}
