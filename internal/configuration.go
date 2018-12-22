package internal

import (
	"github.com/jinzhu/gorm"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/url"
	"time"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

type Config struct {
	LogLevel  string `default:"INFO"`
	LogFormat string `default:"[{{.StartTime}}] {{.Method}} {{.Path}} {{.Status}} {{.Duration}} {{.Request.UserAgent}}"`

	Host string `default:"0.0.0.0"`
	Port int    `default:"80"`

	DBURL string `default:"sqlite3:///etc/speakerbob/speakerbob.db"`
	AuthBackendURL   string `default:"memory://"`
	MessageBrokerURL string `default:"memory://"`
	SoundBackendURL string `default:"local:///etc/speakerbob/sounds"`

	SoundBucketName string `default:"sbsounds"`

	PageSize       int `default:"100"`
	MaxSoundLength int `default:"10"`

	CookieName string        `default:"speakerbob"`
	TokenTTL   time.Duration `default:"172800s"`
}

func GetConfig() *Config {
	var config = &Config{}

	if err := envconfig.Process("", config); err != nil {
		log.Fatalf("failed to parse configuration from environment: %v", err)
	}

	return config
}

func GetDB(dbURL string) *gorm.DB {
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		log.Fatal("invalid database url")
	}

	db, err := gorm.Open(parsedURL.Scheme, parsedURL.Path)

	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}

	return db
}
