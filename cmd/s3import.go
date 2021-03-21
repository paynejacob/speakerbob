package cmd

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var (
	s3Endpoint string
	s3Key      string
	s3Secret   string
	s3Bucket   string
)

func init() {
	s3EndpointFlag := "s3endpoint"
	s3ImportCmd.PersistentFlags().StringVar(&s3Endpoint, s3EndpointFlag, "s3.us-east-2.amazonaws.com", "")
	_ = viper.BindPFlag(s3EndpointFlag, s3ImportCmd.PersistentFlags().Lookup(s3EndpointFlag))

	s3KeyFlag := "s3key"
	s3ImportCmd.PersistentFlags().StringVar(&s3Key, s3KeyFlag, "", "")
	_ = viper.BindPFlag(s3KeyFlag, s3ImportCmd.PersistentFlags().Lookup(s3KeyFlag))
	_ = s3ImportCmd.MarkPersistentFlagRequired(s3KeyFlag)

	s3SecretFlag := "s3secret"
	s3ImportCmd.PersistentFlags().StringVar(&s3Secret, s3SecretFlag, "", "")
	_ = viper.BindPFlag(s3SecretFlag, s3ImportCmd.PersistentFlags().Lookup(s3SecretFlag))
	_ = s3ImportCmd.MarkPersistentFlagRequired(s3SecretFlag)

	s3BucketFlag := "s3bucket"
	s3ImportCmd.PersistentFlags().StringVar(&s3Bucket, s3BucketFlag, "", "")
	_ = viper.BindPFlag(s3BucketFlag, s3ImportCmd.PersistentFlags().Lookup(s3BucketFlag))
	_ = s3ImportCmd.MarkPersistentFlagRequired(s3BucketFlag)

	rootCmd.AddCommand(s3ImportCmd)
}

var s3ImportCmd = &cobra.Command{
	Use:   "s3import",
	Short: "Import sounds from s3.",
	Long:  `Older versions of speakerbob stored sounds in s3. This command will copy them into the local database.`,
	Run:   S3Import,
}

func S3Import(*cobra.Command, []string) {
	minioClient, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Key, s3Secret, ""),
		Secure: true,
	})

	db, err := badger.Open(badger.DefaultOptions(dataPath))
	if err != nil {
		logrus.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	logrus.Info("syncing sounds from s3")

	err = sound.SyncFromS3(db, minioClient, s3Bucket)
	if err != nil {
		log.Fatal(err)
	}

	logrus.Info("done")
}
