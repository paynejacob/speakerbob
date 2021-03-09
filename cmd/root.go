package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	s3Endpoint          string
	s3Key               string
	s3Secret            string
	s3Bucket            string
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
	s3EndpointFlag := "s3endpoint"
	rootCmd.PersistentFlags().StringVar(&s3Endpoint, s3EndpointFlag, "s3.us-east-2.amazonaws.com", "")
	_ = viper.BindPFlag(s3EndpointFlag, rootCmd.PersistentFlags().Lookup(s3EndpointFlag))

	s3KeyFlag := "s3key"
	rootCmd.PersistentFlags().StringVar(&s3Key, s3KeyFlag, "", "")
	_ = viper.BindPFlag(s3KeyFlag, rootCmd.PersistentFlags().Lookup(s3KeyFlag))
	_ = rootCmd.MarkPersistentFlagRequired(s3KeyFlag)

	s3SecretFlag := "s3secret"
	rootCmd.PersistentFlags().StringVar(&s3Secret, s3SecretFlag, "", "")
	_ = viper.BindPFlag(s3SecretFlag, rootCmd.PersistentFlags().Lookup(s3SecretFlag))
	_ = rootCmd.MarkPersistentFlagRequired(s3SecretFlag)

	s3BucketFlag := "s3bucket"
	rootCmd.PersistentFlags().StringVar(&s3Bucket, s3BucketFlag, "", "")
	_ = viper.BindPFlag(s3BucketFlag, rootCmd.PersistentFlags().Lookup(s3BucketFlag))
	_ = rootCmd.MarkPersistentFlagRequired(s3BucketFlag)

	logLevelFlag := "loglevel"
	rootCmd.PersistentFlags().StringVar(&logLevelString, logLevelFlag, "info", "")
	_ = viper.BindPFlag(logLevelFlag, rootCmd.PersistentFlags().Lookup(logLevelFlag))

	durationLimitFlag := "durationlimit"
	rootCmd.PersistentFlags().StringVar(&durationLimitString, durationLimitFlag, "5s", "maximum duration of an uploaded sound.")
	_ = viper.BindPFlag(durationLimitFlag, rootCmd.PersistentFlags().Lookup(durationLimitFlag))
}
