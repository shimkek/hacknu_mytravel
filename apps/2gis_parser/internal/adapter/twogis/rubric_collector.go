package twogis

import (
	"2gis-parser/internal/domain"
	"2gis-parser/internal/helper"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// RubricCollector handles rubric collection and management
type RubricCollector struct {
	api    *API
	logger Logger
}

// NewRubricCollector creates a new rubric collector instance
func NewRubricCollector(api *API, logger Logger) *RubricCollector {
	return &RubricCollector{
		api:    api,
		logger: logger,
	}
}

// SearchBusinesses searches for businesses by keyword in a specific region
func (rc *RubricCollector) SearchBusinesses(keyword, regionID string, pageSize int) (*domain.BusinessSearchResponse, error) {
	// URL encode the keyword
	encodedKeyword := url.QueryEscape(keyword)

	// Construct the API URL
	apiURL := fmt.Sprintf("https://catalog.api.2gis.com/3.0/items?key=%s&region_id=%s&q=%s&page=1&page_size=%d&fields=items.rubrics",
		rc.api.apiKey, regionID, encodedKeyword, pageSize)

	rc.logger.Info("Searching businesses with keyword '%s' in region %s", keyword, regionID)

	resp, err := rc.api.client.Get(apiURL)
	if err != nil {
		rc.logger.Error("Failed to make search API request: %v", err)
		return nil, fmt.Errorf("failed to make search API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		rc.logger.Error("Search API request failed with status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("search API request failed with status code: %d", resp.StatusCode)
	}

	var searchResponse domain.BusinessSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		rc.logger.Error("Failed to decode search API response: %v", err)
		return nil, fmt.Errorf("failed to decode search API response: %w", err)
	}

	rc.logger.Info("Successfully found %d businesses for keyword '%s' (total: %d)",
		len(searchResponse.Result.Items), keyword, searchResponse.Result.Total)

	return &searchResponse, nil
}

// CollectRubricsFromKeywords searches for businesses using multiple keywords and collects all unique rubrics
func (rc *RubricCollector) CollectRubricsFromKeywords(keywords []string, regionID string) ([]domain.RubricData, error) {
	rubricMap := make(map[string]*domain.RubricData)

	rc.logger.Info("Starting rubric collection for %d keywords in region %s", len(keywords), regionID)

	for i, keyword := range keywords {
		rc.logger.Info("Processing keyword %d/%d: %s", i+1, len(keywords), keyword)

		// Search with current keyword
		searchResp, err := rc.SearchBusinesses(keyword, regionID, 50)
		if err != nil {
			rc.logger.Error("Failed to search businesses for keyword '%s': %v", keyword, err)
			continue // Continue with next keyword instead of failing completely
		}

		// Process rubrics from search results
		businessCount := 0
		for _, business := range searchResp.Result.Items {
			businessCount++
			for _, rubric := range business.Rubrics {
				key := rubric.ID
				if existing, exists := rubricMap[key]; exists {
					existing.Count++
				} else {
					rubricMap[key] = &domain.RubricData{
						ID:       rubric.ID,
						Alias:    rubric.Alias,
						Name:     rubric.Name,
						Kind:     rubric.Kind,
						ParentID: rubric.ParentID,
						ShortID:  rubric.ShortID,
						Count:    1,
					}
				}
			}
		}

		rc.logger.Info("Processed %d businesses for keyword '%s'", businessCount, keyword)

		// Add delay to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	// Convert map to slice
	var rubrics []domain.RubricData
	for _, rubric := range rubricMap {
		rubrics = append(rubrics, *rubric)
	}

	rc.logger.Info("Collected %d unique rubrics from %d keywords", len(rubrics), len(keywords))
	return rubrics, nil
}

// SaveRubricsToCSV saves collected rubric data to a CSV file
func (rc *RubricCollector) SaveRubricsToCSV(rubrics []domain.RubricData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Alias", "Name", "Kind", "ParentID", "ShortID", "Count"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rubric data sorted by count (most frequent first)
	for _, rubric := range rc.sortRubricsByCount(rubrics) {
		record := []string{
			rubric.ID,
			rubric.Alias,
			rubric.Name,
			rubric.Kind,
			rubric.ParentID,
			strconv.Itoa(rubric.ShortID),
			strconv.Itoa(rubric.Count),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	rc.logger.Info("Successfully saved %d rubrics to %s", len(rubrics), filename)
	return nil
}

// LoadKeywordsFromCSV loads search keywords from a CSV file
func (rc *RubricCollector) LoadKeywordsFromCSV(filename string) ([]string, error) {
	return helper.LoadKeywordsFromCSV(filename)
}

// GetTopRubrics returns the top N rubrics by frequency
func (rc *RubricCollector) GetTopRubrics(rubrics []domain.RubricData, limit int) []domain.RubricData {
	sorted := rc.sortRubricsByCount(rubrics)
	if len(sorted) > limit {
		return sorted[:limit]
	}
	return sorted
}

// FilterRubricsByKeywords filters rubrics by name containing any of the keywords
func (rc *RubricCollector) FilterRubricsByKeywords(rubrics []domain.RubricData, keywords []string) []domain.RubricData {
	var filtered []domain.RubricData

	for _, rubric := range rubrics {
		rubricNameLower := strings.ToLower(rubric.Name)
		for _, keyword := range keywords {
			if strings.Contains(rubricNameLower, strings.ToLower(keyword)) {
				filtered = append(filtered, rubric)
				break
			}
		}
	}

	return filtered
}

// sortRubricsByCount sorts rubrics by count in descending order
func (rc *RubricCollector) sortRubricsByCount(rubrics []domain.RubricData) []domain.RubricData {
	// Simple bubble sort for small datasets
	sorted := make([]domain.RubricData, len(rubrics))
	copy(sorted, rubrics)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Count < sorted[j+1].Count {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// GenerateRubricReport generates a detailed report of collected rubrics
func (rc *RubricCollector) GenerateRubricReport(rubrics []domain.RubricData) {
	rc.logger.Info("=== RUBRIC COLLECTION REPORT ===")
	rc.logger.Info("Total unique rubrics found: %d", len(rubrics))

	// Count by kind
	kindCount := make(map[string]int)
	for _, rubric := range rubrics {
		kindCount[rubric.Kind]++
	}

	rc.logger.Info("Rubrics by kind:")
	for kind, count := range kindCount {
		rc.logger.Info("  - %s: %d", kind, count)
	}

	// Show top 10 most frequent rubrics
	top10 := rc.GetTopRubrics(rubrics, 10)
	rc.logger.Info("Top 10 most frequent rubrics:")
	for i, rubric := range top10 {
		rc.logger.Info("  %d. %s (%s) - %d occurrences", i+1, rubric.Name, rubric.Alias, rubric.Count)
	}

	// Filter tourism-related rubrics
	tourismKeywords := []string{"отель", "гостин", "турист", "отдых", "база", "кемпинг", "глэмпинг", "санаторий", "курорт", "пансионат"}
	tourismRubrics := rc.FilterRubricsByKeywords(rubrics, tourismKeywords)

	rc.logger.Info("Tourism-related rubrics found: %d", len(tourismRubrics))
	for _, rubric := range tourismRubrics {
		rc.logger.Info("  - %s (%s): %d times", rubric.Name, rubric.Alias, rubric.Count)
	}
}
