package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "" // set at compile time with -ldflags "-X github.com/paynejacob/speakerbob/cmd.Version=x.y.yz"

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
	fmt.Println(version)
}
