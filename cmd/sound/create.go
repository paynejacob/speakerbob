package sound

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/spf13/cobra"
)

var (
	createName      string
	createAudioFile string
)

var createCommand = &cobra.Command{
	Use:           "create",
	Short:         "Create a new sound from an audio file.",
	RunE:          runCreate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	createCommand.Flags().StringVar(&createName, "name", "", "Name for the sound (1-49 characters). If omitted the sound is created hidden, matching the upload API.")
	createCommand.Flags().StringVar(&createAudioFile, "audio-file", "", "Path to the audio file to normalize and store.")
	_ = createCommand.MarkFlagRequired("audio-file")
}

func runCreate(*cobra.Command, []string) error {
	soundProvider, _, closeStore, err := openProviders()
	if err != nil {
		return err
	}
	defer closeStore()

	sound, err := createSound(soundProvider, createName, createAudioFile, durationLimit)
	if err != nil {
		return err
	}

	fmt.Println(sound.Id)
	return nil
}

func createSound(soundProvider *psound.SoundProvider, name, audioFile string, maxDuration time.Duration) (*psound.Sound, error) {
	file, err := os.Open(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file %q: %w", audioFile, err)
	}
	defer file.Close()

	sound, err := soundProvider.NewSound(filepath.Base(audioFile), file, maxDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create sound: %w", err)
	}

	if name != "" {
		if err = setSoundName(soundProvider, sound, name); err != nil {
			return nil, err
		}
	}

	return sound, nil
}
