package usecase

import (
	"2gis-parser/internal/adapter/logger"
	"2gis-parser/internal/domain"
	"2gis-parser/internal/store"
	"fmt"
	"time"
)

type BusinessProvider interface {
	FetchBusinesses(categoryID string, regionID string, fullInfo bool) ([]domain.BusinessDetail, error)
	FetchRegions() ([]domain.Region, error)
	FetchBusinessDetail(id string) (domain.BusinessDetail, error)
}

type Parser struct {
	provider BusinessProvider
	logger   *logger.Logger
	store    *store.PostgresStore
}

func NewParserWithStore(provider BusinessProvider, logger *logger.Logger, dbStore *store.PostgresStore) *Parser {
	return &Parser{
		provider: provider,
		logger:   logger,
		store:    dbStore,
	}
}

func (p *Parser) Run() error {
	p.logger.Info("Starting 2GIS KZ regions and businesses parsing with database integration")

	// Initialize database store if not provided
	if p.store == nil {
		p.logger.Info("Initializing database connection")
		config := store.LoadConfigFromEnv()
		dbStore, err := store.NewPostgresStore(config, p.logger)
		if err != nil {
			p.logger.Error("Failed to connect to database: %v", err)
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer dbStore.Close()
		p.store = dbStore
	}

	startTime := time.Now()
	totalProcessed := 0

	// First, fetch regions from 2GIS API
	regions := []domain.Region{domain.Region{ID: "67", Name: "Almaty"}}
	p.logger.Info("Processing %d regions", len(regions))

	// Load rubrics from filtered CSV file
	// rubrics, err := helper.LoadFieldsFromCSV("docks/filtered_rubricks.csv")
	// if err != nil {
	// 	p.logger.Error("Failed to load rubrics: %v", err)
	// 	return fmt.Errorf("failed to load rubrics: %w", err)
	// }
	rubrics := []string{"70348"}

	p.logger.Info("Loaded %d tourism rubrics from CSV", len(rubrics))

	// Process each region
	for _, region := range regions {
		p.logger.Info("Processing region: %s (%s)", region.Name, region.ID)

		regionProcessed := 0

		// For each region, fetch businesses for each tourism rubric
		for _, rubricID := range rubrics {
			p.logger.Info("Fetching businesses for rubric %s in %s", rubricID, region.Name)

			businesses, err := p.provider.FetchBusinesses(rubricID, region.ID, true)
			if err != nil {
				p.logger.Error("Failed to fetch businesses for rubric %s in region %s: %v", rubricID, region.ID, err)
				continue
			}

			p.logger.Info("Found %d businesses for rubric %s in region %s", len(businesses), rubricID, region.Name)

			if len(businesses) == 0 {
				continue
			}

			// Insert businesses into database - choose between parallel and sequential processing
			if len(businesses) >= 20 { // Use parallel processing for larger batches
				if err := p.store.InsertBusinessDetailsWithBatching(businesses); err != nil {
					p.logger.Error("Failed to process businesses for rubric %s: %v", rubricID, err)
					continue
				}
			} else {
				// Use simple parallel processing for smaller batches
				if err := p.store.InsertBusinessDetails(businesses); err != nil {
					p.logger.Error("Failed to process businesses for rubric %s: %v", rubricID, err)
					continue
				}
			}

			regionProcessed += len(businesses)

			// Log individual business details
			for _, business := range businesses {
				p.logger.Debug("Processed: ID=%s, Name=%s, Address=%s, Rating=%.1f, Type=%s",
					business.ID, business.Name, business.FullAddressName,
					business.Reviews.GeneralRating, p.getAccommodationType(business.Rubrics))
			}

			// Add small delay to respect API rate limits
			time.Sleep(100 * time.Millisecond)
		}

		totalProcessed += regionProcessed
		p.logger.Info("Completed region %s: %d businesses processed", region.Name, regionProcessed)
	}

	// Final summary
	duration := time.Since(startTime)
	p.logger.Info("=== PARSING COMPLETED ===")
	p.logger.Info("Total processed: %d businesses", totalProcessed)
	p.logger.Info("Duration: %v", duration)
	p.logger.Info("Individual business logs stored in parsing_logs table")

	return nil
}

// RunSingleBusiness fetches and stores a single business by ID
func (p *Parser) RunSingleBusiness(businessID string) error {
	if p.store == nil {
		config := store.LoadConfigFromEnv()
		dbStore, err := store.NewPostgresStore(config, p.logger)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer dbStore.Close()
		p.store = dbStore
	}

	p.logger.Info("Fetching single business with ID: %s", businessID)

	business, err := p.provider.FetchBusinessDetail(businessID)
	if err != nil {
		return fmt.Errorf("failed to fetch business detail: %w", err)
	}

	p.logger.Info("Successfully fetched business: %s", business.Name)

	if err := p.store.InsertBusinessDetail(business); err != nil {
		return fmt.Errorf("failed to insert business detail: %w", err)
	}

	p.logger.Info("Successfully stored business in database")
	return nil
}

// getAccommodationType helper function to determine accommodation type for logging
func (p *Parser) getAccommodationType(rubrics []domain.Rubric) string {
	for _, rubric := range rubrics {
		switch rubric.Alias {
		case "gostinicy":
			return "Hotel"
		case "khostely":
			return "Hostel"
		case "bazy_otdykha":
			return "Resort"
		case "arenda_kottedzhejj":
			return "Villa"
		case "kempingi":
			return "Camping"
		case "sanatorii":
			return "Sanatorium"
		}
	}
	return "Other"
}
