package storage

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Dir string `envconfig:"DIR" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("STORAGE", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	return config
}
