package twogis

import (
	"2gis-parser/internal/domain"
	"encoding/json"
	"fmt"
	"net/http"
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

func (a *API) FetchBusinesses(category string, region string) ([]domain.Business, error) {
	return nil, nil
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
