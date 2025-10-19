package main

import (
	"fmt"
	"hacknu/internal/config"
	"hacknu/internal/logger"
	"hacknu/internal/store"
	"hacknu/parser"
	"time"
)

const (
	// Updated URL for Almaty (destination ID: -2335204)
	summaryURL = "https://www.booking.com/searchresults.html?dest_id=-2335204&dest_type=city&nflt=ht_id%3D213%3Bht_id%3D220%3Bht_id%3D214%3Bht_id%3D216"
)

func main() {
	fmt.Println("📍 Starting Booking.com parser for Almaty with database integration")

	// Initialize logger
	logger := logger.New()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database store
	dbStore, err := store.NewPostgresStore(cfg.DB, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer dbStore.Close()

	logger.Info("Database connection established successfully")

	startTime := time.Now()

	// Step 1. Fetch summary list (main search page)
	properties, err := parser.FetchSummary(summaryURL)
	if err != nil {
		logger.Fatal("Failed to fetch summary: %v", err)
	}

	logger.Info("Found %d properties from summary page", len(properties))

	if len(properties) == 0 {
		logger.Fatal("No properties found. Exiting.")
	}

	var finalData []store.BookingProperty
	successCount := 0
	errorCount := 0

	// Step 2. Iterate and fetch details for each property with retries
	for i, prop := range properties {
		logger.Info("[%d/%d] Processing property: %s", i+1, len(properties), prop.PropertyName)

		url := fmt.Sprintf("https://www.booking.com/hotel/kz/%s.html", prop.PageName)

		var detail *parser.DetailedProperty
		var err error

		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			detail, err = parser.ExtractPropertyDetails(url, prop.Description, prop.ReviewsRatings, prop.ReviewsCount)
			if err == nil && detail != nil {
				break
			}

			logger.Warn("Attempt %d/%d failed for %s: %v", attempt, maxRetries, url, err)
			if attempt < maxRetries {
				backoff := time.Duration(2*attempt) * time.Second
				logger.Info("Retrying in %v...", backoff)
				time.Sleep(backoff)
			}
		}

		if err != nil || detail == nil {
			logger.Error("Skipping after %d attempts: %s", maxRetries, url)
			errorCount++
			continue
		}

		// Convert DetailedProperty to BookingProperty for database storage
		bookingProperty := convertToBookingProperty(*detail)
		finalData = append(finalData, bookingProperty)
		successCount++

		// Add delay between requests to be respectful to booking.com
		time.Sleep(3 * time.Second)
	}

	logger.Info("Parsing completed: %d successful, %d failed out of %d total", successCount, errorCount, len(properties))

	// Step 3. Save parsed data to database
	if len(finalData) > 0 {
		logger.Info("Saving %d properties to database...", len(finalData))
		if err := dbStore.InsertBookingProperties(finalData); err != nil {
			logger.Error("Failed to save properties to database: %v", err)
		}

		// Also save to JSON file as backup
		if err := parser.SaveFinalData(convertToDetailedProperties(finalData)); err != nil {
			logger.Error("Failed to save backup JSON file: %v", err)
		} else {
			logger.Info("Backup JSON file saved successfully")
		}

		duration := time.Since(startTime)
		logger.Info("=== PARSING COMPLETED ===")
		logger.Info("Total processed: %d properties", len(finalData))
		logger.Info("Duration: %v", duration)
		logger.Info("Data saved to database and backup JSON file")
	} else {
		logger.Warn("No valid property details were parsed.")
	}
}

// convertToBookingProperty converts parser.DetailedProperty to store.BookingProperty
func convertToBookingProperty(detail parser.DetailedProperty) store.BookingProperty {
	// Convert parser.Review to store.BookingReview
	var reviews []store.BookingReview
	for _, review := range detail.Reviews {
		reviews = append(reviews, store.BookingReview{
			Name:  review.Name,
			Score: review.Score,
		})
	}

	return store.BookingProperty{
		PropertyName:      detail.PropertyName,
		PageName:          detail.PageName,
		URL:               detail.URL,
		Latitude:          detail.Latitude,
		Longitude:         detail.Longitude,
		Address:           detail.Address,
		AccommodationType: detail.AccommodationType,
		Description:       detail.Description,
		Photos:            detail.Photos,
		ReviewsRatings:    detail.ReviewsRatings,
		ReviewsCount:      detail.ReviewsCount,
		Reviews:           reviews,
		Facilities:        detail.Facilities,
		InspectionStatus:  detail.InspectionStatus,
		LastUpdated:       detail.LastUpdated,
	}
}

// convertToDetailedProperties converts store.BookingProperty back to parser.DetailedProperty for JSON backup
func convertToDetailedProperties(bookingProps []store.BookingProperty) []parser.DetailedProperty {
	var detailedProps []parser.DetailedProperty

	for _, prop := range bookingProps {
		// Convert store.BookingReview to parser.Review
		var reviews []parser.Review
		for _, review := range prop.Reviews {
			reviews = append(reviews, parser.Review{
				Name:  review.Name,
				Score: review.Score,
			})
		}

		detailedProps = append(detailedProps, parser.DetailedProperty{
			PropertyName:      prop.PropertyName,
			PageName:          prop.PageName,
			URL:               prop.URL,
			Latitude:          prop.Latitude,
			Longitude:         prop.Longitude,
			Address:           prop.Address,
			AccommodationType: prop.AccommodationType,
			Description:       prop.Description,
			Photos:            prop.Photos,
			ReviewsRatings:    prop.ReviewsRatings,
			ReviewsCount:      prop.ReviewsCount,
			Reviews:           reviews,
			Facilities:        prop.Facilities,
			InspectionStatus:  prop.InspectionStatus,
			LastUpdated:       prop.LastUpdated,
		})
	}

	return detailedProps
}
