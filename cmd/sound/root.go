package sound

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/paynejacob/speakerbob/cmd/server"
	psound "github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/store/badgerdb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	dataPath      string
	durationLimit time.Duration
)

var Command = &cobra.Command{
	Use:   "sound",
	Short: "Manage speakerbob sounds directly against the database.",
}

func init() {
	Command.PersistentFlags().StringVar(&dataPath, "data-path", server.DefaultConfiguration.DataPath, "Path to the speakerbob data directory.")
	Command.PersistentFlags().DurationVar(&durationLimit, "duration-limit", server.DefaultConfiguration.DurationLimit, "Maximum duration for normalized sound audio.")

	Command.AddCommand(createCommand)
	Command.AddCommand(updateCommand)
	Command.AddCommand(deleteCommand)
	Command.AddCommand(normalizeCommand)
}

// openProviders opens the on-disk data store and returns initialized sound
// and group providers plus a close function for the caller to defer.
func openProviders() (soundProvider *psound.SoundProvider, groupProvider *psound.GroupProvider, closeStore func() error, err error) {
	badgerdbOptions := badger.DefaultOptions(dataPath)
	badgerdbOptions.Logger = logrus.StandardLogger()

	db, err := badger.Open(badgerdbOptions)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open data store at %q: %w", dataPath, err)
	}

	_store := badgerdb.Store{DB: db}

	soundProvider = &psound.SoundProvider{Store: _store}
	if err = soundProvider.Initialize(); err != nil {
		_ = _store.Close()
		return nil, nil, nil, fmt.Errorf("failed to initialize sound provider: %w", err)
	}

	groupProvider = &psound.GroupProvider{Store: _store}
	if err = groupProvider.Initialize(); err != nil {
		_ = _store.Close()
		return nil, nil, nil, fmt.Errorf("failed to initialize group provider: %w", err)
	}

	return soundProvider, groupProvider, _store.Close, nil
}

// setSoundName applies this codebase's existing name-length rule (see
// pkg/sound/service.go's updateSound handler) and un-hides the sound, so CLI
// naming behaves identically to the HTTP API.
func setSoundName(soundProvider *psound.SoundProvider, sound *psound.Sound, name string) error {
	if !(0 < len(name) && len(name) < 50) {
		return fmt.Errorf("sound name must be between 1 and 49 characters")
	}

	sound.Name = name
	sound.Hidden = false

	if err := soundProvider.Save(sound); err != nil {
		return fmt.Errorf("failed to save sound %q: %w", sound.Id, err)
	}

	return nil
}
