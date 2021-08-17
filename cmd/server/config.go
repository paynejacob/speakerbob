package server

import (
	"github.com/paynejacob/speakerbob/pkg/auth"
	github "github.com/paynejacob/speakerbob/pkg/auth/github"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Configuration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DataPath string `yaml:"data_path"`

	DurationLimit time.Duration `yaml:"duration_limit"`

	Auth struct {
		Github github.Provider `yaml:"github"`
	} `yaml:"auth"`

	providers []auth.Provider
}

var DefaultConfiguration = Configuration{
	Host:          "0.0.0.0",
	Port:          80,
	DataPath:      "/etc/speakerbob/data",
	DurationLimit: 10 * time.Second,
}

func (c Configuration) Providers() []auth.Provider {
	if c.Auth.Github.Enabled {
		c.providers = append(c.providers, c.Auth.Github)
	}

	return c.providers
}

func parseConfiguration(configFilePath string) (cfg Configuration, err error) {
	var f *os.File

	if configFilePath == "" {
		return DefaultConfiguration, err
	}

	f, err = os.Open(configFilePath)
	if err != nil && !os.IsNotExist(err) {
		return

	}

	// create file if not found
	if os.IsNotExist(err) {
		f, err = os.Create(configFilePath)
		if err != nil {
			return
		}

		err = yaml.NewEncoder(f).Encode(DefaultConfiguration)
		if err != nil {
			return
		}

		_ = f.Close()

		return DefaultConfiguration, nil
	}

	err = yaml.NewDecoder(f).Decode(&cfg)

	return
}
