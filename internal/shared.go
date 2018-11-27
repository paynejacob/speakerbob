package internal

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kelseyhightower/envconfig"
	"log"
	"sync"
)

type Config struct {
	LogLevel  string `default:"INFO"`
	LogFormat string `default:"[{{.StartTime}}] {{.Method}} {{.Path}} {{.Status}} {{.Duration}} {{.Request.UserAgent}}"`

	Host string `default:"0.0.0.0"`
	Port int    `default:"80"`

	DBHost     string
	DBPort     uint `default:"5432"`
	DBUser     string
	DBPassword string
	DBName     string

	RedisURL string
}

var configOnce sync.Once
var config Config

func GetConfig() Config {
	configOnce.Do(func() {
		err := envconfig.Process("", &config)
		if err != nil {
			log.Fatalf("Failed to parse configuration from environment: %v", err)
		}
	})

	return config
}

var dbOnce sync.Once
var db *gorm.DB

func GetDB() *gorm.DB {
	dbOnce.Do(func() {
		var err error
		db, err = gorm.Open(
			"postgres",
			fmt.Sprintf(
				"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
				GetConfig().DBHost,
				GetConfig().DBPort,
				GetConfig().DBUser,
				GetConfig().DBName,
				GetConfig().DBPassword))
		if err != nil {
			log.Fatalf("Failed to connect to %s", err)
		}
	})

	return db
}

var redisClientOnce sync.Once
var redisClient *redis.Client

func GetRedisClient() *redis.Client {
	redisClientOnce.Do(func() {
		var redisOptions, _ = redis.ParseURL(GetConfig().RedisURL)
		redisClient = redis.NewClient(redisOptions)
	})

	return redisClient
}
