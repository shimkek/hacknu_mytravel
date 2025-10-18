package main

import (
	"2gis-parser/internal/adapter/logger"
	"2gis-parser/internal/adapter/twogis"
	"fmt"
	"net/http"
	"time"
)

// Simple test to verify regions parsing works
func testRegionsParsing() {
	l := logger.New("")

	// Use the provided API key from the URL
	apiKey := "401b0774-dbe8-4f70-91c9-697adac3e650"

	client := &http.Client{Timeout: 30 * time.Second}
	api := twogis.NewAPI(client, apiKey, l)

	l.Info("Testing 2GIS regions API...")

	regions, err := api.FetchRegions()
	if err != nil {
		l.Error("Failed to fetch regions: %v", err)
		return
	}

	l.Info("Successfully fetched %d regions", len(regions))

	// Print first few regions for verification
	for i, region := range regions {
		if i < 5 { // Print first 5 regions
			fmt.Printf("Region %d: ID=%s, Name=%s, Type=%s\n", i+1, region.ID, region.Name, region.Type)
		}
	}

	l.Info("Regions parsing test completed successfully")
}

func main() {
	// Run test first
	testRegionsParsing()

	// Original main logic would continue here...
}
