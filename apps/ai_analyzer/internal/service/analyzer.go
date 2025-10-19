package service

import (
	"ai_analyzer/internal/models"
	"ai_analyzer/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AnalyzerService struct {
	repo      *repository.AccommodationRepository
	aiService *OpenAIService
}

func NewAnalyzerService(repo *repository.AccommodationRepository, aiService *OpenAIService) *AnalyzerService {
	return &AnalyzerService{
		repo:      repo,
		aiService: aiService,
	}
}

func (s *AnalyzerService) AnalyzeAccommodations(ctx context.Context, req models.AnalysisRequest) (*models.AnalysisResponse, error) {
	startTime := time.Now()

	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 100
	}

	// Get accommodations from database
	accommodations, err := s.repo.GetAll(req.Filters, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get accommodations: %w", err)
	}

	// Get database stats
	stats, err := s.repo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Prepare data for AI analysis
	dataForAI := s.prepareDataForAI(accommodations, stats)

	// Create prompt for OpenAI
	prompt := s.buildAnalysisPrompt(req.Prompt, dataForAI)

	// Get analysis from OpenAI
	analysis, err := s.aiService.GetAnalysis(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI analysis: %w", err)
	}

	// Create data summary
	dataSummary := s.createDataSummary(accommodations, stats)

	processingTime := time.Since(startTime)

	return &models.AnalysisResponse{
		Analysis:       analysis,
		DataSummary:    dataSummary,
		RecordCount:    len(accommodations),
		RequestedAt:    startTime,
		ProcessingTime: processingTime.String(),
	}, nil
}

func (s *AnalyzerService) AnalyzeSingleAccommodation(ctx context.Context, id int, prompt string) (*models.AnalysisResponse, error) {
	startTime := time.Now()

	// Get single accommodation
	accommodation, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get accommodation: %w", err)
	}

	// Prepare data for AI analysis
	dataForAI := map[string]interface{}{
		"accommodation": accommodation,
	}

	// Create prompt for OpenAI
	fullPrompt := s.buildSingleAnalysisPrompt(prompt, dataForAI)

	// Get analysis from OpenAI
	analysis, err := s.aiService.GetAnalysis(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI analysis: %w", err)
	}

	processingTime := time.Since(startTime)

	return &models.AnalysisResponse{
		Analysis:       analysis,
		DataSummary:    fmt.Sprintf("Analysis of accommodation: %s (ID: %d)", accommodation.Name, accommodation.ID),
		RecordCount:    1,
		RequestedAt:    startTime,
		ProcessingTime: processingTime.String(),
	}, nil
}

func (s *AnalyzerService) GenerateTravelDescription(ctx context.Context, id int) (*models.AnalysisResponse, error) {
	startTime := time.Now()

	// Get single accommodation
	accommodation, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get accommodation: %w", err)
	}

	// Build travel description prompt in Russian
	prompt := fmt.Sprintf(`Ты — профессиональный копирайтер для сайта о путешествиях в Казахстане (mytravel.kz).
Создай SEO-оптимизированное, увлекательное описание туристического объекта (150–300 слов).

Используй данные:
- Название: %s
- Тип размещения: %s
- Локация / адрес: %s
- Инфраструктура: %s
- Отзывы и рейтинг: %v
- Фото (пример): %s
- Особенности: %s

⚙️ Требования:
1. Напиши живо и образно, в стиле travel-журнала.
2. Сделай текст привлекательным для SEO (упомяни "Казахстан", "отдых", "туризм").
3. Включи преимущества, атмосферу и ключевые удобства.
4. Без форматирования — обычный текст, без Markdown.`,
		accommodation.Name,
		getStringValue(accommodation.AccommodationType),
		getStringValue(accommodation.Address),
		getAmenitiesString(accommodation.Amenities),
		accommodation.Rating,
		getFirstPhoto(accommodation.Photos),
		getStringValue(accommodation.ServiceDescription))

	// Get analysis from OpenAI
	description, err := s.aiService.GetAnalysis(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate travel description: %w", err)
	}

	processingTime := time.Since(startTime)

	return &models.AnalysisResponse{
		Analysis:       description,
		DataSummary:    fmt.Sprintf("Travel description generated for: %s (ID: %d)", accommodation.Name, accommodation.ID),
		RecordCount:    1,
		RequestedAt:    startTime,
		ProcessingTime: processingTime.String(),
	}, nil
}

func (s *AnalyzerService) EvaluateAccommodation(ctx context.Context, id int) (*models.AnalysisResponse, error) {
	startTime := time.Now()

	// Get single accommodation
	accommodation, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get accommodation: %w", err)
	}

	// Convert accommodation to JSON for the prompt
	objectData, err := json.MarshalIndent(accommodation, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal accommodation data: %w", err)
	}

	// Build evaluation prompt in Russian with strict JSON output requirement
	prompt := fmt.Sprintf(`Оцени туристический объект по следующим критериям от 1 до 10:

1. Активность в сети — частота постов, подписчики, взаимодействие
2. Полнота данных — наличие контактов, фото, описания
3. Популярность — упоминания, отзывы, рейтинг
4. Потенциал заполняемости — вместимость, цена, расположение
5. Соответствие ЦА — подходит ли премиум, эко, семейному или молодёжному сегменту
6. Коммерческий потенциал — профессионализм, готовность платить за продвижение

ВАЖНО: Верни ТОЛЬКО JSON-объект без дополнительного текста, объяснений или markdown форматирования.

Формат ответа:
{
  "Активность в сети": 7,
  "Полнота данных": 9,
  "Популярность": 8,
  "Потенциал заполняемости": 7,
  "Соответствие ЦА": 8,
  "Коммерческий потенциал": 9,
  "Общий рейтинг": 8.0,
  "Категория": "горячий"
}

Категории приоритета: "горячий" (7-10), "тёплый" (4-6), "холодный" (1-3).

Данные объекта:
%s`, string(objectData))

	// Get analysis from OpenAI
	evaluation, err := s.aiService.GetAnalysis(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate accommodation: %w", err)
	}

	processingTime := time.Since(startTime)

	return &models.AnalysisResponse{
		Analysis:       evaluation,
		DataSummary:    fmt.Sprintf("Priority evaluation for: %s (ID: %d)", accommodation.Name, accommodation.ID),
		RecordCount:    1,
		RequestedAt:    startTime,
		ProcessingTime: processingTime.String(),
	}, nil
}

func (s *AnalyzerService) prepareDataForAI(accommodations []models.Accommodation, stats map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"accommodations": accommodations,
		"stats":          stats,
		"total_records":  len(accommodations),
	}
}

func (s *AnalyzerService) buildAnalysisPrompt(userPrompt string, data map[string]interface{}) string {
	dataJSON, _ := json.MarshalIndent(data, "", "  ")

	return fmt.Sprintf(`You are an expert data analyst specializing in accommodation and tourism data. 
Analyze the following accommodation data and provide insights based on the user's request.

User Request: %s

Data to analyze:
%s

Please provide a comprehensive analysis that includes:
1. Direct answer to the user's question/request
2. Key insights and patterns found in the data
3. Statistical observations
4. Recommendations based on the data
5. Any notable findings or anomalies

Format your response in a clear, structured manner with proper headings and bullet points where appropriate.`,
		userPrompt, string(dataJSON))
}

func (s *AnalyzerService) buildSingleAnalysisPrompt(userPrompt string, data map[string]interface{}) string {
	dataJSON, _ := json.MarshalIndent(data, "", "  ")

	return fmt.Sprintf(`You are an expert accommodation analyst. 
Analyze the following single accommodation and provide insights based on the user's request.

User Request: %s

Accommodation Data:
%s

Please provide a detailed analysis focusing on:
1. Direct answer to the user's question about this accommodation
2. Strengths and weaknesses
3. Market positioning
4. Unique features or characteristics
5. Recommendations for improvement

Format your response clearly and professionally.`,
		userPrompt, string(dataJSON))
}

func (s *AnalyzerService) createDataSummary(accommodations []models.Accommodation, stats map[string]interface{}) string {
	summary := fmt.Sprintf("Analyzed %d accommodations. ", len(accommodations))

	if totalCount, ok := stats["total_count"]; ok {
		summary += fmt.Sprintf("Total in database: %v. ", totalCount)
	}

	if avgRating, ok := stats["average_rating"]; ok {
		summary += fmt.Sprintf("Average rating: %.2f. ", avgRating)
	}

	if sourceStats, ok := stats["by_source"].(map[string]int); ok {
		summary += "Sources: "
		for source, count := range sourceStats {
			summary += fmt.Sprintf("%s (%d), ", source, count)
		}
	}

	return summary
}

// Helper function to get first photo or return placeholder
func getFirstPhoto(photos models.JSONB) string {
	if len(photos) == 0 {
		return "нет"
	}

	// Unmarshal the JSONB data
	data, err := photos.Unmarshal()
	if err != nil {
		return "нет"
	}

	// Handle case where photos is a JSON array directly
	if photoArray, ok := data.([]interface{}); ok && len(photoArray) > 0 {
		if photo, ok := photoArray[0].(string); ok {
			return photo
		}
	}

	// Handle case where photos is a JSON object with a "photos" key
	if photoMap, ok := data.(map[string]interface{}); ok {
		if photoList, ok := photoMap["photos"].([]interface{}); ok && len(photoList) > 0 {
			if photo, ok := photoList[0].(string); ok {
				return photo
			}
		}
	}

	return "нет"
}

// Helper function to safely get string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return "не указано"
	}
	return *ptr
}

// Helper function to convert amenities JSONB to string
func getAmenitiesString(amenities models.JSONB) string {
	if len(amenities) == 0 {
		return "не указано"
	}

	// Return the raw JSON string for amenities
	return string(amenities)
}
