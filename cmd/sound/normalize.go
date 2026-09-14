package sound

import (
	"fmt"
	"time"

	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/spf13/cobra"
)

var (
	normalizeId  string
	normalizeAll bool
)

var normalizeCommand = &cobra.Command{
	Use:           "normalize",
	Short:         "Re-normalize the audio of one sound, or every sound with --all.",
	RunE:          runNormalize,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	normalizeCommand.Flags().StringVar(&normalizeId, "id", "", "ID of the sound to re-normalize.")
	normalizeCommand.Flags().BoolVar(&normalizeAll, "all", false, "Re-normalize every sound.")
}

func runNormalize(*cobra.Command, []string) error {
	soundProvider, _, closeStore, err := openProviders()
	if err != nil {
		return err
	}
	defer closeStore()

	return normalizeSounds(soundProvider, normalizeId, normalizeAll, durationLimit)
}

func normalizeSounds(soundProvider *psound.SoundProvider, id string, all bool, maxDuration time.Duration) error {
	if !all && id == "" {
		return fmt.Errorf("either --id or --all must be provided")
	}
	if all && id != "" {
		return fmt.Errorf("--id and --all are mutually exclusive")
	}

	var sounds []*psound.Sound
	if all {
		sounds = soundProvider.List()
	} else {
		sound := soundProvider.Get(id)
		if sound == nil {
			return fmt.Errorf("sound %q not found", id)
		}
		sounds = []*psound.Sound{sound}
	}

	for _, sound := range sounds {
		if err := soundProvider.Renormalize(sound, maxDuration); err != nil {
			return fmt.Errorf("failed to normalize sound %q: %w", sound.Id, err)
		}
	}

	return nil
}
