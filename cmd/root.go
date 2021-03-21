package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	dataPath            string
	durationLimitString string
	logLevelString      string
)

var rootCmd = &cobra.Command{
	Use:   "speakerbob",
	Short: "Speakerbob is a distributed soundboard.",
	Long:  "Speakerbob is a distributed soundboard.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	dataPathFlag := "datapath"
	rootCmd.PersistentFlags().StringVar(&dataPath, dataPathFlag, "/etc/speakerbob/data", "")
	_ = viper.BindPFlag(dataPathFlag, rootCmd.PersistentFlags().Lookup(dataPathFlag))

	logLevelFlag := "loglevel"
	rootCmd.PersistentFlags().StringVar(&logLevelString, logLevelFlag, "info", "")
	_ = viper.BindPFlag(logLevelFlag, rootCmd.PersistentFlags().Lookup(logLevelFlag))

	durationLimitFlag := "durationlimit"
	rootCmd.PersistentFlags().StringVar(&durationLimitString, durationLimitFlag, "5s", "maximum duration of an uploaded sound.")
	_ = viper.BindPFlag(durationLimitFlag, rootCmd.PersistentFlags().Lookup(durationLimitFlag))
}
