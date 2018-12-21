package internal

import (
	"github.com/jinzhu/gorm"
	"github.com/kelseyhightower/envconfig"
	"github.com/minio/minio-go"
	"log"
	"time"
)

type Config struct {
	LogLevel  string `default:"INFO"`
	LogFormat string `default:"[{{.StartTime}}] {{.Method}} {{.Path}} {{.Status}} {{.Duration}} {{.Request.UserAgent}}"`

	Host string `default:"0.0.0.0"`
	Port int    `default:"80"`

	DBDialect string `default:"sqlite3"`
	DBConfig  string `default:"/etc/speakerbob/speakerbob.db"`

	RedisURL string

	AuthBackendURL   string `default:"memory://"`
	MessageBrokerURL string `default:"memory://"`

	MinioURL        string
	MinioAccessID   string `default:""`
	MinioAccessKey  string `default:""`
	MinioUseSSL     bool   `default:"true"`
	SoundBucketName string `default:"sbsounds"`

	PageSize       int `default:"100"`
	MaxSoundLength int `default:"10"`

	CookieName string        `default:"speakerbob"`
	TokenTTL   time.Duration `default:"-1s"`
}

func GetConfig() *Config {
	var config = &Config{}

	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("failed to parse configuration from environment: %v", err)
	}

	return config
}

func GetDB(dialect string, config string) *gorm.DB {
	log.Println("connecting to database")

	db, err := gorm.Open(dialect, config)

	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}

	return db
}

func GetMinio(url string, accessID string, accessKey string, useSSL bool) *minio.Client {
	log.Println("connecting to minio")

	client, err := minio.New(url, accessID, accessKey, useSSL)

	if err != nil {
		log.Fatalf("Failed to config minio %v", err)
	}

	return client
}
