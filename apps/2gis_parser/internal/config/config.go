package config

import (
	"os"
)

type Config struct {
	TwoGisAPIKey string
}

func Load() *Config {
	return &Config{
		TwoGisAPIKey: os.Getenv("TWO_GIS_API_KEY"),
	}
}
