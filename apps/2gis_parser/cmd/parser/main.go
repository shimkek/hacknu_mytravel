package main

import (
	"2gis-parser/internal/adapter/logger"
	"2gis-parser/internal/adapter/twogis"
	"2gis-parser/internal/config"
	"2gis-parser/internal/store"
	"2gis-parser/internal/usecase"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	l := logger.New("development")

	err := godotenv.Load()
	if err != nil {
		l.Info("No .env file found, using environment variables")
	}

	cfg := config.Load()
	if cfg.TwoGisAPIKey == "" {
		l.Fatal("TWO_GIS_API_KEY not found in environment, using default key for testing")
	}

	// Initialize database connection
	dbConfig := store.LoadConfigFromEnv()
	dbStore, err := store.NewPostgresStore(dbConfig, l)
	if err != nil {
		l.Fatal("Failed to connect to database: %v", err)
	}
	defer dbStore.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	api := twogis.NewAPI(client, cfg.TwoGisAPIKey, l)
	parser := usecase.NewParserWithStore(api, l, dbStore)

	// Parse command line flags
	var (
		collectRubrics = flag.Bool("collect-rubrics", false, "Collect rubrics data from tourism keywords")
		regionID       = flag.String("region", "67", "Region ID for rubrics collection (default: 67 for Almaty)")
		keywordsFile   = flag.String("keywords", "docks/tourism_keywords.csv", "Path to keywords CSV file")
		outputFile     = flag.String("output", "rubrics_data.csv", "Output CSV file for rubrics data")
		singleBusiness = flag.String("business", "", "Fetch and store a single business by ID")
		parallel       = flag.Bool("parallel", true, "Use parallel processing for database inserts (default: true)")
		workers        = flag.Int("workers", 5, "Number of parallel workers for database inserts (default: 5)")
	)
	flag.Parse()

	// Configure parallel processing if specified
	if *parallel {
		l.Info("Parallel processing enabled with %d workers", *workers)
	} else {
		l.Info("Sequential processing enabled")
	}

	if *singleBusiness != "" {
		l.Info("Fetching single business with ID: %s", *singleBusiness)
		if err := parser.RunSingleBusiness(*singleBusiness); err != nil {
			l.Fatal("Failed to fetch single business: %v", err)
		}
		return
	}

	if *collectRubrics {
		l.Info("Starting rubrics collection process")
		if err := collectRubricsData(api, l, *regionID, *keywordsFile, *outputFile); err != nil {
			l.Fatal("Failed to collect rubrics: %v", err)
		}
		return
	}

	for {
		if err := parser.Run(); err != nil {
			l.Fatal("failed to run parser: %v", err)
		}
		l.Info("Parser run completed, sleeping for 10 minutes before next run")
		time.Sleep(60 * time.Minute)
	}
}

func collectRubricsData(api *twogis.API, l *logger.Logger, regionID, keywordsFile, outputFile string) error {
	// Create rubric collector
	collector := twogis.NewRubricCollector(api, l)

	// Load keywords from CSV file
	keywords, err := collector.LoadKeywordsFromCSV(keywordsFile)
	if err != nil {
		return fmt.Errorf("failed to load keywords: %w", err)
	}

	l.Info("Loaded %d keywords from %s", len(keywords), keywordsFile)

	// Collect rubrics from keywords
	rubrics, err := collector.CollectRubricsFromKeywords(keywords, regionID)
	if err != nil {
		return fmt.Errorf("failed to collect rubrics: %w", err)
	}

	l.Info("Collected %d unique rubrics", len(rubrics))

	// Generate detailed report
	collector.GenerateRubricReport(rubrics)

	// Save rubrics to CSV file
	if err := collector.SaveRubricsToCSV(rubrics, outputFile); err != nil {
		return fmt.Errorf("failed to save rubrics to CSV: %w", err)
	}

	l.Info("Rubrics data saved to %s", outputFile)

	return nil
}
