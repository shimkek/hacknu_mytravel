package repository

import (
	"ai_analyzer/internal/models"
	"database/sql"
	"fmt"
)

type AccommodationRepository struct {
	db *sql.DB
}

func NewAccommodationRepository(db *sql.DB) *AccommodationRepository {
	return &AccommodationRepository{db: db}
}

func (r *AccommodationRepository) GetAll(filters map[string]interface{}, limit int) ([]models.Accommodation, error) {
	query := `
		SELECT id, name, latitude, longitude, address, phone, email, 
		       social_media_links, website_url, social_media_page, 
		       service_description, room_count, capacity, price_range_min, 
		       price_range_max, price_currency, photos, rating, review_count, 
		       reviews, amenities, verification_status, last_updated, 
		       source_website, source_url, external_id, created_at, 
		       deleted_at, accommodation_type
		FROM accommodations 
		WHERE deleted_at IS NULL`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters != nil {
		if sourceWebsite, ok := filters["source_website"]; ok {
			query += fmt.Sprintf(" AND source_website = $%d", argIndex)
			args = append(args, sourceWebsite)
			argIndex++
		}
		if verificationStatus, ok := filters["verification_status"]; ok {
			query += fmt.Sprintf(" AND verification_status = $%d", argIndex)
			args = append(args, verificationStatus)
			argIndex++
		}
		if minRating, ok := filters["min_rating"]; ok {
			query += fmt.Sprintf(" AND rating >= $%d", argIndex)
			args = append(args, minRating)
			argIndex++
		}
		if accommodationType, ok := filters["accommodation_type"]; ok {
			query += fmt.Sprintf(" AND accommodation_type = $%d", argIndex)
			args = append(args, accommodationType)
			argIndex++
		}
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query accommodations: %w", err)
	}
	defer rows.Close()

	var accommodations []models.Accommodation
	for rows.Next() {
		var acc models.Accommodation
		err := rows.Scan(
			&acc.ID, &acc.Name, &acc.Latitude, &acc.Longitude, &acc.Address,
			&acc.Phone, &acc.Email, &acc.SocialMediaLinks, &acc.WebsiteURL,
			&acc.SocialMediaPage, &acc.ServiceDescription, &acc.RoomCount,
			&acc.Capacity, &acc.PriceRangeMin, &acc.PriceRangeMax,
			&acc.PriceCurrency, &acc.Photos, &acc.Rating, &acc.ReviewCount,
			&acc.Reviews, &acc.Amenities, &acc.VerificationStatus,
			&acc.LastUpdated, &acc.SourceWebsite, &acc.SourceURL,
			&acc.ExternalID, &acc.CreatedAt, &acc.DeletedAt, &acc.AccommodationType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan accommodation: %w", err)
		}
		accommodations = append(accommodations, acc)
	}

	return accommodations, nil
}

func (r *AccommodationRepository) GetByID(id int) (*models.Accommodation, error) {
	query := `
		SELECT id, name, latitude, longitude, address, phone, email, 
		       social_media_links, website_url, social_media_page, 
		       service_description, room_count, capacity, price_range_min, 
		       price_range_max, price_currency, photos, rating, review_count, 
		       reviews, amenities, verification_status, last_updated, 
		       source_website, source_url, external_id, created_at, 
		       deleted_at, accommodation_type
		FROM accommodations 
		WHERE id = $1 AND deleted_at IS NULL`

	var acc models.Accommodation
	err := r.db.QueryRow(query, id).Scan(
		&acc.ID, &acc.Name, &acc.Latitude, &acc.Longitude, &acc.Address,
		&acc.Phone, &acc.Email, &acc.SocialMediaLinks, &acc.WebsiteURL,
		&acc.SocialMediaPage, &acc.ServiceDescription, &acc.RoomCount,
		&acc.Capacity, &acc.PriceRangeMin, &acc.PriceRangeMax,
		&acc.PriceCurrency, &acc.Photos, &acc.Rating, &acc.ReviewCount,
		&acc.Reviews, &acc.Amenities, &acc.VerificationStatus,
		&acc.LastUpdated, &acc.SourceWebsite, &acc.SourceURL,
		&acc.ExternalID, &acc.CreatedAt, &acc.DeletedAt, &acc.AccommodationType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get accommodation by ID: %w", err)
	}

	return &acc, nil
}

func (r *AccommodationRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	var totalCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM accommodations WHERE deleted_at IS NULL").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_count"] = totalCount

	// By source
	rows, err := r.db.Query(`
		SELECT source_website, COUNT(*) 
		FROM accommodations 
		WHERE deleted_at IS NULL 
		GROUP BY source_website`)
	if err != nil {
		return nil, fmt.Errorf("failed to get source stats: %w", err)
	}
	defer rows.Close()

	sourceStats := make(map[string]int)
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err == nil {
			sourceStats[source] = count
		}
	}
	stats["by_source"] = sourceStats

	// Average rating
	var avgRating sql.NullFloat64
	err = r.db.QueryRow("SELECT AVG(rating) FROM accommodations WHERE deleted_at IS NULL AND rating IS NOT NULL").Scan(&avgRating)
	if err == nil && avgRating.Valid {
		stats["average_rating"] = avgRating.Float64
	}

	return stats, nil
}
