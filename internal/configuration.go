package internal

import (
	"github.com/IBM-Cloud/bluemix-go"
	"github.com/IBM-Cloud/bluemix-go/session"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // ensure gorm supports mysql
	_ "github.com/jinzhu/gorm/dialects/postgres" // ensure gorm supports pg
	_ "github.com/jinzhu/gorm/dialects/sqlite"  // ensure gorm supports sqlite
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/url"
	"time"
)

type Config struct {
	LogLevel  string `default:"INFO"`
	LogFormat string `default:"[{{.StartTime}}] {{.Method}} {{.Path}} {{.Status}} {{.Duration}} {{.Request.UserAgent}}"`

	Host string `default:"0.0.0.0"`
	Port int    `default:"80"`

	DBURL            string `default:"sqlite3:///etc/speakerbob/speakerbob.db"`
	AuthBackendURL   string `default:""`
	SearchBackendURL string `default:"memory://"`
	MessageBrokerURL string `default:"memory://"`
	SoundBackendURL  string `default:"local:///etc/speakerbob/sounds"`
	BluemixAPIKey    string `default:""`

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

func GetBluemixSession(bluemixAPIKey string) *session.Session {
	if bluemixAPIKey != "" {
		return nil
	}
	sess, err := session.New(&bluemix.Config{BluemixAPIKey: bluemixAPIKey})

	if err != nil {
		log.Fatalf("failed to configure bluemix session: %v", err)
	}

	return sess
}
