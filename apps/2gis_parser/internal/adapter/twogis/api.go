package twogis

import (
	"2gis-parser/internal/domain"
	"2gis-parser/internal/helper"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type API struct {
	client *http.Client
	apiKey string
	logger Logger
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

func NewAPI(client *http.Client, apiKey string, logger Logger) *API {
	return &API{client: client, apiKey: apiKey, logger: logger}
}

func (a *API) FetchBusinesses(categoryID string, regionID string, fullInfo bool) ([]domain.BusinessDetail, error) {
	var allBusinesses []domain.BusinessDetail
	page := 1
	pageSize := 50

	for {
		// Build the base URL
		url := fmt.Sprintf("https://catalog.api.2gis.com/3.0/items?rubric_id=%s&key=%s&region_id=%s", categoryID, a.apiKey, regionID)

		// Add fields if fullInfo is true
		if fullInfo {
			// Load fields from CSV file
			fields, err := helper.LoadFieldsFromCSV("docks/fields.csv")
			if err != nil {
				a.logger.Error("Failed to load fields from CSV: %v", err)
				return nil, fmt.Errorf("failed to load fields from CSV: %w", err)
			}

			if len(fields) > 0 {
				url += "&fields=" + strings.Join(fields, ",")
			}
		}

		// Add pagination parameters
		url += fmt.Sprintf("&page=%d&page_size=%d", page, pageSize)

		// a.logger.Info("Fetching businesses from 2GIS API (page %d): %s", page, url)

		resp, err := a.client.Get(url)
		if err != nil {
			a.logger.Error("Failed to make API request for businesses (page %d): %v", page, err)
			// If it's the first page, return the error, otherwise return what we have
			if page == 1 {
				return nil, fmt.Errorf("failed to make API request: %w", err)
			}
			a.logger.Info("Error on page %d, returning %d businesses collected so far", page, len(allBusinesses))
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			a.logger.Error("API request failed with status code: %d (page %d)", resp.StatusCode, page)
			// If it's the first page, return the error, otherwise return what we have
			if page == 1 {
				return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
			}
			a.logger.Info("Error on page %d, returning %d businesses collected so far", page, len(allBusinesses))
			break
		}

		// Use BusinessByIdResponse since the structure is the same for items endpoint
		var businessResponse domain.BusinessByIdResponse
		if err := json.NewDecoder(resp.Body).Decode(&businessResponse); err != nil {
			a.logger.Error("Failed to decode businesses API response (page %d): %v", page, err)
			// If it's the first page, return the error, otherwise return what we have
			if page == 1 {
				return nil, fmt.Errorf("failed to decode API response: %w", err)
			}
			a.logger.Info("Error decoding page %d, returning %d businesses collected so far", page, len(allBusinesses))
			break
		}

		pageBusinesses := businessResponse.Result.Items
		a.logger.Info("Successfully fetched %d businesses on page %d for rubric ID: %s in region: %s", len(pageBusinesses), page, categoryID, regionID)

		// If no businesses returned, we've reached the end
		if len(pageBusinesses) == 0 {
			a.logger.Info("No more businesses found, reached end of results at page %d", page)
			break
		}

		// Add businesses from this page to our collection
		allBusinesses = append(allBusinesses, pageBusinesses...)
		time.Sleep(100 * time.Millisecond) // Respectful delay between requests
		// If we got fewer businesses than the page size, we've reached the end
		if len(pageBusinesses) < pageSize {
			a.logger.Info("Received %d businesses (less than page size %d), reached end of results", len(pageBusinesses), pageSize)
			break
		}

		page++
	}

	a.logger.Info("Completed pagination for rubric ID: %s in region: %s. Total businesses collected: %d", categoryID, regionID, len(allBusinesses))
	return allBusinesses, nil
}

// FetchBusinessesByRubrics fetches businesses by rubric IDs (original implementation)
func (a *API) FetchBusinessesByRubrics(rubricIDs []string, regionID string) ([]domain.BusinessDetail, error) {
	return nil, nil
}

func (a *API) FetchBusinessDetail(id string) (domain.BusinessDetail, error) {
	// Construct the API URL with all required fields
	url := fmt.Sprintf("https://catalog.api.2gis.com/3.0/items/byid?key=%s&id=%s&fields=items.point,items.full_address_name,items.rubrics,items.schedule,items.description,items.flags,items.reviews,items.statistics,items.dates.updated_at,items.caption,items.stat,items.schedule_special,items.attribute_groups,items.reg_bc_url,items.address,items.links,items.summary", a.apiKey, id)

	a.logger.Info("Fetching business detail from 2GIS API: %s", url)

	resp, err := a.client.Get(url)
	if err != nil {
		a.logger.Error("Failed to make API request for business detail: %v", err)
		return domain.BusinessDetail{}, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.logger.Error("API request failed with status code: %d", resp.StatusCode)
		return domain.BusinessDetail{}, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var businessResponse domain.BusinessByIdResponse
	if err := json.NewDecoder(resp.Body).Decode(&businessResponse); err != nil {
		a.logger.Error("Failed to decode business detail API response: %v", err)
		return domain.BusinessDetail{}, fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(businessResponse.Result.Items) == 0 {
		a.logger.Error("No business found with ID: %s", id)
		return domain.BusinessDetail{}, fmt.Errorf("no business found with ID: %s", id)
	}

	businessDetail := businessResponse.Result.Items[0]
	a.logger.Info("Successfully fetched business detail for ID: %s, Name: %s", id, businessDetail.Name)

	return businessDetail, nil
}

// FetchRegions fetches all available regions from 2GIS API for Kazakhstan
func (a *API) FetchRegions() ([]domain.Region, error) {
	url := fmt.Sprintf("https://catalog.api.2gis.com/2.0/region/list?key=%s&locale=ru_RU&country_code_filter=kz", a.apiKey)

	a.logger.Info("Fetching regions from 2GIS API: %s", url)

	resp, err := a.client.Get(url)
	if err != nil {
		a.logger.Error("Failed to make API request: %v", err)
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.logger.Error("API request failed with status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var regionsResponse domain.RegionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&regionsResponse); err != nil {
		a.logger.Error("Failed to decode API response: %v", err)
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	a.logger.Info("Successfully fetched %d regions (total: %d)", len(regionsResponse.Result.Items), regionsResponse.Result.Total)

	return regionsResponse.Result.Items, nil
}
