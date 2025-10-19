package main

import (
	"ai_analyzer/internal/config"
	"ai_analyzer/internal/handler"
	"ai_analyzer/internal/repository"
	"ai_analyzer/internal/service"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize dependencies
	repo := repository.NewAccommodationRepository(db)
	aiService := service.NewOpenAIService(cfg.OpenAIAPIKey)
	analyzerService := service.NewAnalyzerService(repo, aiService)
	handler := handler.NewAnalyzerHandler(analyzerService)

	// Setup routes
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// AI analysis endpoints
	api := r.Group("/api/v1")
	{
		api.POST("/analyze", handler.AnalyzeAccommodations)
		api.GET("/accommodations", handler.GetAccommodations)
		api.GET("/accommodations/:id", handler.GetAccommodationByID)
		api.POST("/analyze/accommodation/:id", handler.AnalyzeSingleAccommodation)

		// New convenient endpoints that don't require request body
		// ЭТИ ЭНДПОИНТЫ ЮЗАЙТЕ ДЛЯ ЗАПРОСОВ, ТОЛЬКО АЙДИ АККОМОДЕЙШЕНА ИЗ ДБ ПРОСИТ
		api.POST("/accommodation/:id/description", handler.GenerateTravelDescription)
		api.POST("/accommodation/:id/evaluate", handler.EvaluateAccommodation)
	}

	log.Printf("AI Analyzer service starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
