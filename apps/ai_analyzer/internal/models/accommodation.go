package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Accommodation struct {
	ID                 int        `json:"id" db:"id"`
	Name               string     `json:"name" db:"name"`
	Latitude           *float64   `json:"latitude" db:"latitude"`
	Longitude          *float64   `json:"longitude" db:"longitude"`
	Address            *string    `json:"address" db:"address"`
	Phone              *string    `json:"phone" db:"phone"`
	Email              *string    `json:"email" db:"email"`
	SocialMediaLinks   JSONB      `json:"social_media_links" db:"social_media_links"`
	WebsiteURL         *string    `json:"website_url" db:"website_url"`
	SocialMediaPage    *string    `json:"social_media_page" db:"social_media_page"`
	ServiceDescription *string    `json:"service_description" db:"service_description"`
	RoomCount          *int       `json:"room_count" db:"room_count"`
	Capacity           *int       `json:"capacity" db:"capacity"`
	PriceRangeMin      *float64   `json:"price_range_min" db:"price_range_min"`
	PriceRangeMax      *float64   `json:"price_range_max" db:"price_range_max"`
	PriceCurrency      string     `json:"price_currency" db:"price_currency"`
	Photos             JSONB      `json:"photos" db:"photos"`
	Rating             *float64   `json:"rating" db:"rating"`
	ReviewCount        int        `json:"review_count" db:"review_count"`
	Reviews            JSONB      `json:"reviews" db:"reviews"`
	Amenities          JSONB      `json:"amenities" db:"amenities"`
	VerificationStatus string     `json:"verification_status" db:"verification_status"`
	LastUpdated        time.Time  `json:"last_updated" db:"last_updated"`
	SourceWebsite      string     `json:"source_website" db:"source_website"`
	SourceURL          *string    `json:"source_url" db:"source_url"`
	ExternalID         *string    `json:"external_id" db:"external_id"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	DeletedAt          *time.Time `json:"deleted_at" db:"deleted_at"`
	AccommodationType  *string    `json:"accommodation_type" db:"accommodation_type"`
}

type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	*j = JSONB(bytes)
	return nil
}

// Helper method to unmarshal JSONB to interface{}
func (j JSONB) Unmarshal() (interface{}, error) {
	if len(j) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(j, &result); err != nil {
		return nil, err
	}

	return result, nil
}

type AnalysisRequest struct {
	Prompt  string                 `json:"prompt" binding:"required"`
	Filters map[string]interface{} `json:"filters,omitempty"`
	Limit   int                    `json:"limit,omitempty"`
}

type AnalysisResponse struct {
	Analysis       string    `json:"analysis"`
	DataSummary    string    `json:"data_summary"`
	RecordCount    int       `json:"record_count"`
	RequestedAt    time.Time `json:"requested_at"`
	ProcessingTime string    `json:"processing_time"`
}
