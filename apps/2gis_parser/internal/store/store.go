package store

import (
	"2gis-parser/internal/adapter/logger"
	"2gis-parser/internal/domain"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db     *sql.DB
	logger *logger.Logger
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresStore creates a new PostgreSQL store instance with connection pooling
func NewPostgresStore(config DatabaseConfig, logger *logger.Logger) (*PostgresStore, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pooling for parallel operations
	db.SetMaxOpenConns(10)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database with connection pooling")

	return &PostgresStore{
		db:     db,
		logger: logger,
	}, nil
}

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5433"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "mytravel_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// Close closes the database connection
func (ps *PostgresStore) Close() error {
	return ps.db.Close()
}

// InsertBusinessDetail inserts a BusinessDetail from 2GIS API into accommodations table
func (ps *PostgresStore) InsertBusinessDetail(business domain.BusinessDetail) error {
	startTime := time.Now()

	ps.logger.Debug("Processing business ID: %s, Name: %s", business.ID, business.Name)

	// Convert BusinessDetail to accommodation record
	accommodation := ps.convertBusinessDetailToAccommodation(business)

	// Debug logging - log the JSON data being generated
	ps.debugLogJSONFields(business.ID, accommodation)

	// Validate JSON data before insertion
	if err := ps.validateJSONFields(accommodation); err != nil {
		ps.logger.Error("Invalid JSON data for business %s: %v", business.ID, err)
		ps.logBusinessInsertion("2gis", "failed", fmt.Sprintf("Invalid JSON data: %v", err), startTime)
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	query := `
		INSERT INTO accommodations (
			name, latitude, longitude, address, accommodation_type,
			phone, email, social_media_links, website_url, social_media_page,
			service_description, room_count, capacity, price_range_min, price_range_max,
			photos, rating, review_count, reviews, amenities,
			verification_status, source_website, source_url, external_id
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24
		)
		ON CONFLICT (source_website, external_id) 
		DO UPDATE SET
			name = EXCLUDED.name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			address = EXCLUDED.address,
			accommodation_type = EXCLUDED.accommodation_type,
			phone = EXCLUDED.phone,
			email = EXCLUDED.email,
			social_media_links = EXCLUDED.social_media_links,
			website_url = EXCLUDED.website_url,
			social_media_page = EXCLUDED.social_media_page,
			service_description = EXCLUDED.service_description,
			room_count = EXCLUDED.room_count,
			capacity = EXCLUDED.capacity,
			price_range_min = EXCLUDED.price_range_min,
			price_range_max = EXCLUDED.price_range_max,
			photos = EXCLUDED.photos,
			rating = EXCLUDED.rating,
			review_count = EXCLUDED.review_count,
			reviews = EXCLUDED.reviews,
			amenities = EXCLUDED.amenities,
			verification_status = EXCLUDED.verification_status,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := ps.db.Exec(query,
		accommodation.Name,
		accommodation.Latitude,
		accommodation.Longitude,
		accommodation.Address,
		accommodation.AccommodationType,
		accommodation.Phone,
		accommodation.Email,
		ps.safeJSONBytes(accommodation.SocialMediaLinks),
		accommodation.WebsiteURL,
		accommodation.SocialMediaPage,
		accommodation.ServiceDescription,
		accommodation.RoomCount,
		accommodation.Capacity,
		accommodation.PriceRangeMin,
		accommodation.PriceRangeMax,
		ps.safeJSONBytes(accommodation.Photos),
		accommodation.Rating,
		accommodation.ReviewCount,
		ps.safeJSONBytes(accommodation.Reviews),
		ps.safeJSONBytes(accommodation.Amenities),
		accommodation.VerificationStatus,
		accommodation.SourceWebsite,
		accommodation.SourceURL,
		accommodation.ExternalID,
	)

	if err != nil {
		ps.logger.Error("Failed to insert business detail %s: %v", business.ID, err)
		ps.logBusinessInsertion("2gis", "failed", fmt.Sprintf("Database error: %v", err), startTime)
		return fmt.Errorf("failed to insert business detail: %w", err)
	}

	ps.logger.Info("Successfully inserted/updated accommodation: %s (ID: %s)", business.Name, business.ID)
	ps.logBusinessInsertion("2gis", "success", "", startTime)
	return nil
}

// InsertBusinessDetails inserts multiple BusinessDetail entities in parallel
func (ps *PostgresStore) InsertBusinessDetails(businesses []domain.BusinessDetail) error {
	if len(businesses) == 0 {
		return nil
	}

	// Configuration for parallel processing
	const maxWorkers = 5 // Limit concurrent connections to avoid overwhelming the database
	const batchSize = 10 // Process businesses in small batches for better error handling

	successCount := 0
	errorCount := 0

	// Create channels for work distribution
	businessChan := make(chan domain.BusinessDetail, len(businesses))
	resultChan := make(chan insertResult, len(businesses))

	// Start worker goroutines
	for i := 0; i < maxWorkers; i++ {
		go ps.insertWorker(businessChan, resultChan)
	}

	// Send businesses to workers
	for _, business := range businesses {
		businessChan <- business
	}
	close(businessChan)

	// Collect results
	for i := 0; i < len(businesses); i++ {
		result := <-resultChan
		if result.err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	ps.logger.Info("Parallel processing completed: %d successful, %d failed out of %d total",
		successCount, errorCount, len(businesses))

	return nil
}

// insertResult represents the result of a single business insertion
type insertResult struct {
	businessID string
	err        error
}

// insertWorker processes businesses from the channel
func (ps *PostgresStore) insertWorker(businessChan <-chan domain.BusinessDetail, resultChan chan<- insertResult) {
	for business := range businessChan {
		err := ps.InsertBusinessDetail(business)
		resultChan <- insertResult{
			businessID: business.ID,
			err:        err,
		}
	}
}

// InsertBusinessDetailsWithBatching inserts businesses in parallel batches for better performance
func (ps *PostgresStore) InsertBusinessDetailsWithBatching(businesses []domain.BusinessDetail) error {
	if len(businesses) == 0 {
		return nil
	}

	const batchSize = 20
	const maxConcurrentBatches = 3

	var batches [][]domain.BusinessDetail
	for i := 0; i < len(businesses); i += batchSize {
		end := i + batchSize
		if end > len(businesses) {
			end = len(businesses)
		}
		batches = append(batches, businesses[i:end])
	}

	ps.logger.Info("Processing %d businesses in %d batches with %d concurrent workers",
		len(businesses), len(batches), maxConcurrentBatches)

	// Use semaphore to limit concurrent batches
	semaphore := make(chan struct{}, maxConcurrentBatches)
	var wg sync.WaitGroup

	totalSuccess := 0
	totalErrors := 0
	var mu sync.Mutex

	for i, batch := range batches {
		wg.Add(1)
		go func(batchIndex int, businessBatch []domain.BusinessDetail) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			ps.logger.Info("Starting batch %d with %d businesses", batchIndex+1, len(businessBatch))

			batchSuccess := 0
			batchErrors := 0

			for _, business := range businessBatch {
				if err := ps.InsertBusinessDetail(business); err != nil {
					batchErrors++
				} else {
					batchSuccess++
				}
			}

			// Update totals safely
			mu.Lock()
			totalSuccess += batchSuccess
			totalErrors += batchErrors
			mu.Unlock()

			ps.logger.Info("Completed batch %d: %d successful, %d failed",
				batchIndex+1, batchSuccess, batchErrors)
		}(i, batch)
	}

	wg.Wait()

	ps.logger.Info("All batches completed: %d successful, %d failed out of %d total",
		totalSuccess, totalErrors, len(businesses))

	return nil
}

// logBusinessInsertion logs individual business insertion operations
func (ps *PostgresStore) logBusinessInsertion(sourceWebsite, status, errorMessage string, startTime time.Time) {
	completedAt := time.Now()
	duration := int(completedAt.Sub(startTime).Seconds())

	var errMsg *string
	if errorMessage != "" {
		errMsg = &errorMessage
	}

	query := `
		INSERT INTO parsing_logs (
			source_website, status, error_message, started_at, completed_at, duration_seconds
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := ps.db.Exec(query,
		sourceWebsite,
		status,
		errMsg,
		startTime,
		completedAt,
		duration,
	)

	if err != nil {
		ps.logger.Error("Failed to log business insertion: %v", err)
	}
}

// LogParsingActivity is deprecated - use logBusinessInsertion instead
func (ps *PostgresStore) LogParsingActivity(sourceWebsite string, status string, recordsProcessed, recordsInserted, recordsUpdated int, errorMessage *string) error {
	// This method is kept for compatibility but should not be used for new logging
	ps.logger.Info("LogParsingActivity is deprecated, individual business logs are used instead")
	return nil
}

// validateJSONFields validates JSON fields before database insertion
func (ps *PostgresStore) validateJSONFields(accommodation AccommodationRecord) error {
	// Validate SocialMediaLinks JSON
	if accommodation.SocialMediaLinks != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.SocialMediaLinks, &temp); err != nil {
			return fmt.Errorf("invalid social_media_links JSON: %w", err)
		}
	}

	// Validate Photos JSON
	if accommodation.Photos != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.Photos, &temp); err != nil {
			return fmt.Errorf("invalid photos JSON: %w", err)
		}
	}

	// Validate Reviews JSON
	if accommodation.Reviews != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.Reviews, &temp); err != nil {
			return fmt.Errorf("invalid reviews JSON: %w", err)
		}
	}

	// Validate Amenities JSON
	if accommodation.Amenities != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.Amenities, &temp); err != nil {
			return fmt.Errorf("invalid amenities JSON: %w", err)
		}
	}

	return nil
}

// convertBusinessDetailToAccommodation converts 2GIS BusinessDetail to accommodation record
func (ps *PostgresStore) convertBusinessDetailToAccommodation(business domain.BusinessDetail) AccommodationRecord {
	// Extract phone from attributes if available
	phone := ps.extractPhoneFromAttributes(business.AttributeGroups)

	// Extract price range from attributes
	priceMin, priceMax := ps.extractPriceRange(business.AttributeGroups)

	// Extract capacity from attributes
	capacity := ps.extractCapacity(business.AttributeGroups)

	// Convert amenities from attribute groups
	amenities := ps.convertAttributesToAmenities(business.AttributeGroups)

	// Convert reviews to JSONB
	reviewsJSON := ps.convertReviewsToJSON(business.Reviews)

	// Determine accommodation type from rubrics
	accommodationType := ps.determineAccommodationType(business.Rubrics)

	// Generate photos JSON (placeholder - would need actual photo URLs from API)
	photosJSON := ps.generatePhotosJSON(business.Flags.Photos)

	return AccommodationRecord{
		Name:               business.Name,
		Latitude:           &business.Point.Lat,
		Longitude:          &business.Point.Lon,
		Address:            &business.FullAddressName,
		AccommodationType:  &accommodationType,
		Phone:              phone,
		Email:              nil, // Not available in 2GIS API
		SocialMediaLinks:   nil, // Not available in current API response
		WebsiteURL:         nil, // Not available in current API response
		SocialMediaPage:    nil, // Not available in current API response
		ServiceDescription: ps.generateServiceDescription(business),
		RoomCount:          nil, // Not available in 2GIS API
		Capacity:           capacity,
		PriceRangeMin:      priceMin,
		PriceRangeMax:      priceMax,
		Photos:             photosJSON,
		Rating:             ps.convertRating(business.Reviews.GeneralRating),
		ReviewCount:        &business.Reviews.GeneralReviewCount,
		Reviews:            reviewsJSON,
		Amenities:          amenities,
		VerificationStatus: "new",
		SourceWebsite:      "2gis",
		SourceURL:          nil, // Could be constructed from business ID
		ExternalID:         business.ID,
	}
}

// Helper methods for conversion

func (ps *PostgresStore) extractPhoneFromAttributes(attributeGroups []domain.AttributeGroup) *string {
	// Phone numbers are typically not exposed in public 2GIS API
	// This is a placeholder for when contact_groups become available
	return nil
}

func (ps *PostgresStore) extractPriceRange(attributeGroups []domain.AttributeGroup) (*float64, *float64) {
	for _, group := range attributeGroups {
		for _, attr := range group.Attributes {
			if strings.Contains(attr.Tag, "price") || strings.Contains(attr.Tag, "prozhot") {
				// Extract price from name like "Цена от 18000 тнг."
				if strings.Contains(attr.Name, "от") && strings.Contains(attr.Name, "тнг") {
					// Simple price extraction - could be improved with regex
					priceStr := strings.Replace(attr.Name, "Цена от ", "", 1)
					priceStr = strings.Replace(priceStr, " тнг.", "", 1)
					priceStr = strings.Replace(priceStr, " ", "", -1)

					var price float64
					if _, err := fmt.Sscanf(priceStr, "%f", &price); err == nil {
						return &price, nil
					}
				}
			}
		}
	}
	return nil, nil
}

func (ps *PostgresStore) extractCapacity(attributeGroups []domain.AttributeGroup) *int {
	for _, group := range attributeGroups {
		for _, attr := range group.Attributes {
			if strings.Contains(attr.Tag, "capacity") {
				// Extract capacity from name like "До 120 мест"
				if strings.Contains(attr.Name, "До") && strings.Contains(attr.Name, "мест") {
					var capacity int
					if _, err := fmt.Sscanf(attr.Name, "До %d мест", &capacity); err == nil {
						return &capacity
					}
				}
			}
		}
	}
	return nil
}

// convertAttributesToAmenities converts attribute groups to amenities JSON
func (ps *PostgresStore) convertAttributesToAmenities(attributeGroups []domain.AttributeGroup) []byte {
	amenities := make(map[string]bool)

	for _, group := range attributeGroups {
		for _, attr := range group.Attributes {
			cleanName := ps.sanitizeString(attr.Name)
			switch {
			case strings.Contains(strings.ToLower(cleanName), "завтрак"):
				amenities["breakfast"] = true
			case strings.Contains(strings.ToLower(cleanName), "кафе"):
				amenities["restaurant"] = true
			case strings.Contains(strings.ToLower(cleanName), "беседки"):
				amenities["gazebo"] = true
			case strings.Contains(strings.ToLower(cleanName), "детская площадка"):
				amenities["playground"] = true
			case strings.Contains(strings.ToLower(cleanName), "наличный"):
				amenities["cash_payment"] = true
			case strings.Contains(strings.ToLower(cleanName), "qr"):
				amenities["qr_payment"] = true
			}
		}
	}

	if len(amenities) == 0 {
		return nil
	}

	jsonData, err := json.Marshal(amenities)
	if err != nil {
		ps.logger.Error("Failed to marshal amenities JSON: %v", err)
		return nil
	}
	return jsonData
}

// convertReviewsToJSON converts review info to JSON with proper error handling
func (ps *PostgresStore) convertReviewsToJSON(reviewInfo domain.ReviewInfo) []byte {
	// Create a safe map avoiding potential NaN or Inf values
	reviews := map[string]interface{}{
		"general_rating":       ps.sanitizeFloat(reviewInfo.GeneralRating),
		"general_review_count": reviewInfo.GeneralReviewCount,
		"org_rating":           ps.sanitizeFloat(reviewInfo.OrgRating),
		"org_review_count":     reviewInfo.OrgReviewCount,
		"is_reviewable":        reviewInfo.IsReviewable,
	}

	jsonData, err := json.Marshal(reviews)
	if err != nil {
		ps.logger.Error("Failed to marshal reviews JSON: %v", err)
		return nil
	}
	return jsonData
}

// sanitizeFloat ensures float values are valid for JSON
func (ps *PostgresStore) sanitizeFloat(f float64) float64 {
	// Check for NaN
	if f != f {
		return 0.0
	}
	// Check for positive and negative infinity
	if f > 1.7976931348623157e+308 || f < -1.7976931348623157e+308 {
		return 0.0
	}
	// Check for infinity using math comparison (safer than division)
	if f == f+1 { // This is true for infinity
		return 0.0
	}
	return f
}

// generatePhotosJSON creates photos JSON with proper error handling
func (ps *PostgresStore) generatePhotosJSON(hasPhotos bool) []byte {
	if !hasPhotos {
		return nil
	}

	// Create empty array instead of placeholder to avoid issues
	photos := []string{}
	jsonData, err := json.Marshal(photos)
	if err != nil {
		ps.logger.Error("Failed to marshal photos JSON: %v", err)
		return nil
	}
	return jsonData
}

// generateServiceDescription creates service description with safe string handling
func (ps *PostgresStore) generateServiceDescription(business domain.BusinessDetail) *string {
	var descriptions []string

	// Safely add caption
	if business.Caption != "" {
		descriptions = append(descriptions, ps.sanitizeString(business.Caption))
	}

	// Add rubric information
	for _, rubric := range business.Rubrics {
		if rubric.Kind == "primary" && rubric.Name != "" {
			descriptions = append(descriptions, ps.sanitizeString(rubric.Name))
		}
	}

	// Add key amenities (limit to avoid overly long descriptions)
	count := 0
	for _, group := range business.AttributeGroups {
		if count >= 3 { // Limit to 3 attribute groups
			break
		}
		if group.IsPrimary && len(group.Attributes) > 0 && group.Name != "" && group.Attributes[0].Name != "" {
			descriptions = append(descriptions, fmt.Sprintf("%s: %s",
				ps.sanitizeString(group.Name), ps.sanitizeString(group.Attributes[0].Name)))
			count++
		}
	}

	if len(descriptions) == 0 {
		return nil
	}

	description := strings.Join(descriptions, ". ")
	// Limit description length to avoid database issues
	if len(description) > 1000 {
		description = description[:997] + "..."
	}
	return &description
}

// sanitizeString removes potentially problematic characters from strings
func (ps *PostgresStore) sanitizeString(s string) string {
	// Remove null bytes and other control characters that might cause issues
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\u0000", "")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	// Remove any other problematic Unicode control characters
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r != 127 { // Keep printable characters except DEL
			result.WriteRune(r)
		} else {
			result.WriteRune(' ') // Replace control chars with space
		}
	}

	return strings.TrimSpace(result.String())
}

// safeJSONBytes ensures JSON bytes are valid or returns null
func (ps *PostgresStore) safeJSONBytes(data []byte) interface{} {
	if data == nil {
		return nil
	}

	// Validate that it's actually valid JSON
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		ps.logger.Error("Invalid JSON bytes detected, returning null: %v", err)
		return nil
	}

	return string(data)
}

// debugLogJSONFields logs the JSON fields for debugging purposes
func (ps *PostgresStore) debugLogJSONFields(businessID string, accommodation AccommodationRecord) {
	ps.logger.Debug("=== JSON Debug for business %s ===", businessID)

	if accommodation.SocialMediaLinks != nil {
		ps.logger.Debug("SocialMediaLinks JSON: %s", string(accommodation.SocialMediaLinks))
		var temp interface{}
		if err := json.Unmarshal(accommodation.SocialMediaLinks, &temp); err != nil {
			ps.logger.Error("INVALID SocialMediaLinks JSON for %s: %v", businessID, err)
		}
	}

	if accommodation.Photos != nil {
		ps.logger.Debug("Photos JSON: %s", string(accommodation.Photos))
		var temp interface{}
		if err := json.Unmarshal(accommodation.Photos, &temp); err != nil {
			ps.logger.Error("INVALID Photos JSON for %s: %v", businessID, err)
		}
	}

	if accommodation.Reviews != nil {
		ps.logger.Debug("Reviews JSON: %s", string(accommodation.Reviews))
		var temp interface{}
		if err := json.Unmarshal(accommodation.Reviews, &temp); err != nil {
			ps.logger.Error("INVALID Reviews JSON for %s: %v", businessID, err)
		}
	}

	if accommodation.Amenities != nil {
		ps.logger.Debug("Amenities JSON: %s", string(accommodation.Amenities))
		var temp interface{}
		if err := json.Unmarshal(accommodation.Amenities, &temp); err != nil {
			ps.logger.Error("INVALID Amenities JSON for %s: %v", businessID, err)
		}
	}
}

// determineAccommodationType determines the accommodation type from rubrics
func (ps *PostgresStore) determineAccommodationType(rubrics []domain.Rubric) string {
	for _, rubric := range rubrics {
		switch rubric.Alias {
		case "gostinicy":
			return "hotel"
		case "khostely":
			return "hostel"
		case "bazy_otdykha":
			return "resort"
		case "arenda_kottedzhejj":
			return "villa"
		case "kempingi":
			return "camping"
		case "sanatorii":
			return "resort"
		}
	}
	return "other"
}

// convertRating converts and sanitizes rating values
func (ps *PostgresStore) convertRating(rating float64) *float64 {
	sanitized := ps.sanitizeFloat(rating)
	if sanitized == 0 {
		return nil
	}
	return &sanitized
}

// AccommodationRecord represents the accommodation data for database insertion
type AccommodationRecord struct {
	Name               string
	Latitude           *float64
	Longitude          *float64
	Address            *string
	AccommodationType  *string
	Phone              *string
	Email              *string
	SocialMediaLinks   []byte
	WebsiteURL         *string
	SocialMediaPage    *string
	ServiceDescription *string
	RoomCount          *int
	Capacity           *int
	PriceRangeMin      *float64
	PriceRangeMax      *float64
	Photos             []byte
	Rating             *float64
	ReviewCount        *int
	Reviews            []byte
	Amenities          []byte
	VerificationStatus string
	SourceWebsite      string
	SourceURL          *string
	ExternalID         string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
