package usecase

import (
	"2gis-parser/internal/domain"
	"time"
)

type BusinessProvider interface {
	FetchBusinesses(category string, region string) ([]domain.Business, error)
	FetchRegions() ([]domain.Region, error)
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type Parser struct {
	provider BusinessProvider
	logger   Logger
}

func NewParser(provider BusinessProvider, logger Logger) *Parser {
	return &Parser{provider: provider, logger: logger}
}

func (p *Parser) Run() error {
	p.logger.Info("Starting 2GIS regions parsing")

	for {
		// Fetch regions from 2GIS API
		regions, err := p.provider.FetchRegions()
		if err != nil {
			p.logger.Error("Failed to fetch regions: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}

		p.logger.Info("Processing %d regions", len(regions))

		// Process each region
		for _, region := range regions {
			p.logger.Info("Processing region - ID: %s, Name: %s, Type: %s",
				region.ID, region.Name, region.Type)

			// Here you can add logic to:
			// 1. Store regions in database
			// 2. Use regions to fetch businesses for each region
			// 3. Process region-specific data
		}

		p.logger.Info("Completed regions parsing cycle, waiting 30 minutes before next run")
		time.Sleep(30 * time.Minute) // Run every 30 minutes
	}
}
