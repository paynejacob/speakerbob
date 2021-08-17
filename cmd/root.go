package cmd

import (
	"fmt"
	"github.com/paynejacob/speakerbob/cmd/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	logLevelString string
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
	logLevelFlag := "loglevel"
	rootCmd.PersistentFlags().StringVar(&logLevelString, logLevelFlag, "info", "")

	rootCmd.AddCommand(server.Command)

	level, err := logrus.ParseLevel(logLevelString)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)
}
