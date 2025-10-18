package yandexparser
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"mytravel/internal/database"
	"mytravel/internal/models"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize database connection
	config := database.LoadConfig()
	db, err := database.Connect(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Yandex Parser started")

	// Simple counter that runs every 35 seconds
	counter := 0
	for {
		counter++
		log.Printf("Yandex Parser iteration #%d", counter)

		// Create a simple parsing log entry
		_, err := db.Exec(`
			INSERT INTO parsing_logs (source_website, status, records_processed, started_at) 
			VALUES ($1, $2, $3, $4)`,
			models.SourceWebsiteYandex, "completed", counter, time.Now())

		if err != nil {
			log.Printf("Error inserting parsing log: %v", err)
		} else {
			log.Printf("Successfully logged parsing iteration #%d", counter)
		}

		// Simple accommodation entry (for testing)
		if counter%7 == 0 { // Every 7th iteration, add a test accommodation
			_, err := db.Exec(`
				INSERT INTO accommodations (name, description, source_website, external_id, verification_status, created_at) 
				VALUES ($1, $2, $3, $4, $5, $6)`,
				fmt.Sprintf("Yandex Maps Location #%d", counter),
				fmt.Sprintf("An accommodation found via Yandex Maps, iteration %d", counter),
				models.SourceWebsiteYandex,
				fmt.Sprintf("yandex_%d", counter),
				models.VerificationStatusNew,
				time.Now())

			if err != nil {
				log.Printf("Error inserting test accommodation: %v", err)
			} else {
				log.Printf("Successfully created test accommodation #%d", counter)
			}
		}

		time.Sleep(35 * time.Second)
	}
}