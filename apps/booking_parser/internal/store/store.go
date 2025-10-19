package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hacknu/internal/config"
	"hacknu/internal/logger"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewPostgresStore creates a new PostgreSQL store instance with connection pooling
func NewPostgresStore(cfg config.DatabaseConfig, logger *logger.Logger) (*PostgresStore, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pooling
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database")

	return &PostgresStore{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (ps *PostgresStore) Close() error {
	return ps.db.Close()
}

// BookingProperty represents the structure from booking parser
type BookingProperty struct {
	PropertyName      string          `json:"property_name"`
	PageName          string          `json:"page_name"`
	URL               string          `json:"url"`
	Latitude          float64         `json:"latitude"`
	Longitude         float64         `json:"longitude"`
	Address           string          `json:"address"`
	AccommodationType string          `json:"accommodation_type"`
	Description       string          `json:"description"`
	Photos            []string        `json:"photos"`
	ReviewsRatings    string          `json:"reviews_ratings"`
	ReviewsCount      int             `json:"reviews_count"`
	Reviews           []BookingReview `json:"reviews"`
	Facilities        []string        `json:"facilities"`
	InspectionStatus  string          `json:"inspection_status"`
	LastUpdated       string          `json:"last_updated"`
}

// BookingReview represents individual review from booking
type BookingReview struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
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

// InsertBookingProperty inserts a BookingProperty into the accommodations table
func (ps *PostgresStore) InsertBookingProperty(property BookingProperty) error {
	startTime := time.Now()

	ps.logger.Debug("Processing booking property: %s, Address: %s", property.PropertyName, property.Address)

	// Convert BookingProperty to accommodation record
	accommodation := ps.convertBookingPropertyToAccommodation(property)

	// Validate JSON data before insertion
	if err := ps.validateJSONFields(accommodation); err != nil {
		ps.logger.Error("Invalid JSON data for booking property %s: %v", property.PageName, err)
		ps.logBusinessInsertion("booking", property.PageName, "insert", "failed", fmt.Sprintf("Invalid JSON data: %v", err), startTime)
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	// Perform insert or update
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
		RETURNING (xmax = 0) AS was_insert
	`

	var wasInsert bool
	err := ps.db.QueryRow(query,
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
		"booking",
		accommodation.SourceURL,
		accommodation.ExternalID,
	).Scan(&wasInsert)

	if err != nil {
		ps.logger.Error("Failed to insert/update booking property %s: %v", property.PageName, err)
		operation := "insert"
		ps.logBusinessInsertion("booking", property.PageName, operation, "failed", fmt.Sprintf("Database error: %v", err), startTime)
		return fmt.Errorf("failed to insert/update booking property: %w", err)
	}

	operation := "update"
	if wasInsert {
		operation = "insert"
	}

	ps.logger.Info("Successfully %sed accommodation: %s (PageName: %s)", operation, property.PropertyName, property.PageName)
	ps.logBusinessInsertion("booking", property.PageName, operation, "success", "", startTime)
	return nil
}

// InsertBookingProperties inserts multiple BookingProperty entities in parallel
func (ps *PostgresStore) InsertBookingProperties(properties []BookingProperty) error {
	if len(properties) == 0 {
		return nil
	}

	successCount := 0
	errorCount := 0

	for _, property := range properties {
		if err := ps.InsertBookingProperty(property); err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	ps.logger.Info("Booking processing completed: %d successful, %d failed out of %d total",
		successCount, errorCount, len(properties))

	return nil
}

// convertBookingPropertyToAccommodation converts Booking.com property to accommodation record
func (ps *PostgresStore) convertBookingPropertyToAccommodation(property BookingProperty) AccommodationRecord {
	// Convert photos to JSON
	photosJSON := ps.convertBookingPhotosToJSON(property.Photos)

	// Convert reviews to JSON
	reviewsJSON := ps.convertBookingReviewsToJSON(property.Reviews, property.ReviewsRatings, property.ReviewsCount)

	// Convert facilities to amenities JSON
	amenitiesJSON := ps.convertBookingFacilitiesToAmenities(property.Facilities)

	// Extract rating from reviews_ratings string
	rating := ps.parseBookingRating(property.ReviewsRatings)

	// Generate website URL from page name
	var websiteURL *string
	if property.PageName != "" {
		url := fmt.Sprintf("https://www.booking.com/hotel/kz/%s.html", property.PageName)
		websiteURL = &url
	}

	return AccommodationRecord{
		Name:               property.PropertyName,
		Latitude:           ps.safeFloat64Pointer(property.Latitude),
		Longitude:          ps.safeFloat64Pointer(property.Longitude),
		Address:            ps.safeStringPointer(property.Address),
		AccommodationType:  ps.safeStringPointer(property.AccommodationType),
		Phone:              nil,
		Email:              nil,
		SocialMediaLinks:   nil,
		WebsiteURL:         websiteURL,
		SocialMediaPage:    nil,
		ServiceDescription: ps.safeStringPointer(property.Description),
		RoomCount:          nil,
		Capacity:           nil,
		PriceRangeMin:      nil,
		PriceRangeMax:      nil,
		Photos:             photosJSON,
		Rating:             rating,
		ReviewCount:        ps.safeIntPointer(property.ReviewsCount),
		Reviews:            reviewsJSON,
		Amenities:          amenitiesJSON,
		VerificationStatus: "new", // Changed from property.InspectionStatus to "new"
		SourceWebsite:      "booking",
		SourceURL:          websiteURL,
		ExternalID:         property.PageName,
	}
}

// Helper methods

func (ps *PostgresStore) convertBookingPhotosToJSON(photos []string) []byte {
	if len(photos) == 0 {
		return nil
	}

	jsonData, err := json.Marshal(photos)
	if err != nil {
		ps.logger.Error("Failed to marshal booking photos JSON: %v", err)
		return nil
	}
	return jsonData
}

func (ps *PostgresStore) convertBookingReviewsToJSON(reviews []BookingReview, ratingsStr string, count int) []byte {
	reviewData := map[string]interface{}{
		"general_rating":       ps.parseBookingRatingFloat(ratingsStr),
		"general_review_count": count,
		"detailed_reviews":     reviews,
		"source":               "booking",
	}

	jsonData, err := json.Marshal(reviewData)
	if err != nil {
		ps.logger.Error("Failed to marshal booking reviews JSON: %v", err)
		return nil
	}
	return jsonData
}

func (ps *PostgresStore) convertBookingFacilitiesToAmenities(facilities []string) []byte {
	if len(facilities) == 0 {
		return nil
	}

	amenities := make(map[string]bool)

	for _, facility := range facilities {
		cleanFacility := strings.ToLower(ps.sanitizeString(facility))
		switch {
		case strings.Contains(cleanFacility, "wifi") || strings.Contains(cleanFacility, "internet"):
			amenities["wifi"] = true
		case strings.Contains(cleanFacility, "parking"):
			amenities["parking"] = true
		case strings.Contains(cleanFacility, "pool") || strings.Contains(cleanFacility, "swimming"):
			amenities["pool"] = true
		case strings.Contains(cleanFacility, "gym") || strings.Contains(cleanFacility, "fitness"):
			amenities["gym"] = true
		case strings.Contains(cleanFacility, "spa") || strings.Contains(cleanFacility, "wellness"):
			amenities["spa"] = true
		case strings.Contains(cleanFacility, "restaurant") || strings.Contains(cleanFacility, "dining"):
			amenities["restaurant"] = true
		case strings.Contains(cleanFacility, "bar") || strings.Contains(cleanFacility, "lounge"):
			amenities["bar"] = true
		case strings.Contains(cleanFacility, "breakfast"):
			amenities["breakfast"] = true
		case strings.Contains(cleanFacility, "room service"):
			amenities["room_service"] = true
		case strings.Contains(cleanFacility, "laundry"):
			amenities["laundry"] = true
		case strings.Contains(cleanFacility, "air conditioning") || strings.Contains(cleanFacility, "a/c"):
			amenities["ac"] = true
		case strings.Contains(cleanFacility, "heating"):
			amenities["heating"] = true
		case strings.Contains(cleanFacility, "tv") || strings.Contains(cleanFacility, "television"):
			amenities["tv"] = true
		case strings.Contains(cleanFacility, "minibar"):
			amenities["minibar"] = true
		case strings.Contains(cleanFacility, "safe"):
			amenities["safe"] = true
		case strings.Contains(cleanFacility, "balcony") || strings.Contains(cleanFacility, "terrace"):
			amenities["balcony"] = true
		case strings.Contains(cleanFacility, "kitchen") || strings.Contains(cleanFacility, "kitchenette"):
			amenities["kitchen"] = true
		case strings.Contains(cleanFacility, "pet") || strings.Contains(cleanFacility, "animal"):
			amenities["pets_allowed"] = true
		case strings.Contains(cleanFacility, "wheelchair") || strings.Contains(cleanFacility, "accessible"):
			amenities["disabled_access"] = true
		}
	}

	if len(amenities) == 0 {
		return nil
	}

	jsonData, err := json.Marshal(amenities)
	if err != nil {
		ps.logger.Error("Failed to marshal booking amenities JSON: %v", err)
		return nil
	}
	return jsonData
}

func (ps *PostgresStore) parseBookingRating(ratingsStr string) *float64 {
	if ratingsStr == "" {
		return nil
	}

	var rating float64
	if _, err := fmt.Sscanf(ratingsStr, "%f", &rating); err == nil {
		if rating >= 0 && rating <= 10 {
			return &rating
		}
	}
	return nil
}

func (ps *PostgresStore) parseBookingRatingFloat(ratingsStr string) float64 {
	if rating := ps.parseBookingRating(ratingsStr); rating != nil {
		return *rating
	}
	return 0.0
}

func (ps *PostgresStore) safeFloat64Pointer(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func (ps *PostgresStore) safeStringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (ps *PostgresStore) safeIntPointer(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func (ps *PostgresStore) sanitizeString(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\u0000", "")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r != 127 {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}

	return strings.TrimSpace(result.String())
}

func (ps *PostgresStore) safeJSONBytes(data []byte) interface{} {
	if data == nil {
		return nil
	}

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		ps.logger.Error("Invalid JSON bytes detected, returning null: %v", err)
		return nil
	}

	return string(data)
}

func (ps *PostgresStore) validateJSONFields(accommodation AccommodationRecord) error {
	if accommodation.SocialMediaLinks != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.SocialMediaLinks, &temp); err != nil {
			return fmt.Errorf("invalid social_media_links JSON: %w", err)
		}
	}

	if accommodation.Photos != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.Photos, &temp); err != nil {
			return fmt.Errorf("invalid photos JSON: %w", err)
		}
	}

	if accommodation.Reviews != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.Reviews, &temp); err != nil {
			return fmt.Errorf("invalid reviews JSON: %w", err)
		}
	}

	if accommodation.Amenities != nil {
		var temp interface{}
		if err := json.Unmarshal(accommodation.Amenities, &temp); err != nil {
			return fmt.Errorf("invalid amenities JSON: %w", err)
		}
	}

	return nil
}

func (ps *PostgresStore) logBusinessInsertion(sourceWebsite, externalID, operation, status, errorMessage string, startTime time.Time) {
	completedAt := time.Now()
	duration := int(completedAt.Sub(startTime).Milliseconds())

	var errMsg *string
	if errorMessage != "" {
		errMsg = &errorMessage
	}

	query := `
		INSERT INTO parsing_logs (
			source_website, operation, status, error_message, started_at, completed_at, duration_ms, external_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := ps.db.Exec(query,
		sourceWebsite,
		operation,
		status,
		errMsg,
		startTime,
		completedAt,
		duration,
		externalID,
	)

	if err != nil {
		ps.logger.Error("Failed to log business %s: %v", operation, err)
	}
}
