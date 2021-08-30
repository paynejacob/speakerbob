package cmd

import (
	"fmt"
	"github.com/paynejacob/speakerbob/pkg/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Output the speakerbob version.",
	Long:  `Output the speakerbob version.`,
	Run:   Version,
}

func Version(*cobra.Command, []string) {
	fmt.Println(version.Version)
}
