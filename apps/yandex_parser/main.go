package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database config from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5434")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "mytravel_db")

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Yandex Parser started successfully")

	// Simple counter that runs every 40 seconds
	counter := 0
	for {
		counter++
		log.Printf("Yandex Parser iteration #%d", counter)

		// Log parsing activity
		_, err := db.Exec(`
			INSERT INTO parsing_logs (source_website, status, records_processed, started_at) 
			VALUES ($1, $2, $3, $4)`,
			"yandex", "completed", counter, time.Now())

		if err != nil {
			log.Printf("Error inserting parsing log: %v", err)
		} else {
			log.Printf("Successfully logged parsing iteration #%d", counter)
		}

		// Add test accommodation every 3rd iteration
		if counter%3 == 0 {
			_, err := db.Exec(`
				INSERT INTO accommodations (name, service_description, source_website, external_id, verification_status, created_at) 
				VALUES ($1, $2, $3, $4, $5, $6)`,
				fmt.Sprintf("Yandex Отель #%d", counter),
				fmt.Sprintf("Тестовый отель найденный парсером Yandex на итерации %d", counter),
				"yandex",
				fmt.Sprintf("yandex_%d", counter),
				"new",
				time.Now())

			if err != nil {
				log.Printf("Error inserting test accommodation: %v", err)
			} else {
				log.Printf("Successfully created test accommodation #%d", counter)
			}
		}

		time.Sleep(40 * time.Second)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
