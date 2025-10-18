package instagramparser
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

	log.Println("Instagram Parser started")

	// Simple counter that runs every 60 seconds
	counter := 0
	for {
		counter++
		log.Printf("Instagram Parser iteration #%d", counter)

		// Create a simple parsing log entry
		_, err := db.Exec(`
			INSERT INTO parsing_logs (source_website, status, records_processed, started_at) 
			VALUES ($1, $2, $3, $4)`,
			models.SourceWebsiteInstagram, "completed", counter, time.Now())

		if err != nil {
			log.Printf("Error inserting parsing log: %v", err)
		} else {
			log.Printf("Successfully logged parsing iteration #%d", counter)
		}

		// Simple accommodation entry (for testing)
		if counter%4 == 0 { // Every 4th iteration, add a test accommodation
			_, err := db.Exec(`
				INSERT INTO accommodations (name, description, source_website, external_id, verification_status, created_at) 
				VALUES ($1, $2, $3, $4, $5, $6)`,
				fmt.Sprintf("Instagram Featured Place #%d", counter),
				fmt.Sprintf("A trendy accommodation found on Instagram, iteration %d", counter),
				models.SourceWebsiteInstagram,
				fmt.Sprintf("insta_%d", counter),
				models.VerificationStatusNew,
				time.Now())

			if err != nil {
				log.Printf("Error inserting test accommodation: %v", err)
			} else {
				log.Printf("Successfully created test accommodation #%d", counter)
			}
		}

		time.Sleep(60 * time.Second)
	}
}