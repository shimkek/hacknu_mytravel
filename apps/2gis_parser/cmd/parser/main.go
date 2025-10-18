package main

import (
	"2gis-parser/internal/adapter/logger"
	"2gis-parser/internal/adapter/twogis"
	"2gis-parser/internal/config"
	"2gis-parser/internal/usecase"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	l := logger.New("")

	err := godotenv.Load()
	if err != nil {
		l.Info("No .env file found, using environment variables")
	}

	cfg := config.Load()
	if cfg.TwoGisAPIKey == "" {
		l.Fatal("TWO_GIS_API_KEY not found in environment, using default key for testing")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	api := twogis.NewAPI(client, cfg.TwoGisAPIKey, l)
	parser := usecase.NewParser(api, l)

	l.Info("Starting 2gis regions parser")
	if err := parser.Run(); err != nil {
		l.Fatal("failed to run parser: %v", err)
	}
}
