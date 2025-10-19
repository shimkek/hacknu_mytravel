package handler

import (
	"ai_analyzer/internal/models"
	"ai_analyzer/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnalyzerHandler struct {
	analyzerService *service.AnalyzerService
}

func NewAnalyzerHandler(analyzerService *service.AnalyzerService) *AnalyzerHandler {
	return &AnalyzerHandler{
		analyzerService: analyzerService,
	}
}

func (h *AnalyzerHandler) AnalyzeAccommodations(c *gin.Context) {
	var req models.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.analyzerService.AnalyzeAccommodations(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AnalyzerHandler) GetAccommodations(c *gin.Context) {
	// Parse query parameters for filters
	filters := make(map[string]interface{})

	if sourceWebsite := c.Query("source_website"); sourceWebsite != "" {
		filters["source_website"] = sourceWebsite
	}
	if verificationStatus := c.Query("verification_status"); verificationStatus != "" {
		filters["verification_status"] = verificationStatus
	}
	if minRatingStr := c.Query("min_rating"); minRatingStr != "" {
		if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
			filters["min_rating"] = minRating
		}
	}
	if accommodationType := c.Query("accommodation_type"); accommodationType != "" {
		filters["accommodation_type"] = accommodationType
	}

	// Parse limit
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	// This would need to be implemented in the analyzer service
	// For now, we'll return a simple response
	c.JSON(http.StatusOK, gin.H{
		"message": "Use POST /api/v1/analyze to get AI analysis of accommodations",
		"filters": filters,
		"limit":   limit,
	})
}

func (h *AnalyzerHandler) GetAccommodationByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid accommodation ID"})
		return
	}

	// This would need to be implemented to return raw accommodation data
	// For now, we'll return a simple response
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Use POST /api/v1/analyze/accommodation/%d to get AI analysis", id),
		"id":      id,
	})
}

func (h *AnalyzerHandler) AnalyzeSingleAccommodation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid accommodation ID"})
		return
	}

	var req struct {
		Prompt string `json:"prompt" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.analyzerService.AnalyzeSingleAccommodation(c.Request.Context(), id, req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Generate SEO travel description for accommodation
func (h *AnalyzerHandler) GenerateTravelDescription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid accommodation ID"})
		return
	}

	response, err := h.analyzerService.GenerateTravelDescription(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Evaluate accommodation with priority scoring
func (h *AnalyzerHandler) EvaluateAccommodation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid accommodation ID"})
		return
	}

	response, err := h.analyzerService.EvaluateAccommodation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
