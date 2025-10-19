package service

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client *openai.Client
}

func NewOpenAIService(apiKey string) *OpenAIService {
	client := openai.NewClient(apiKey)
	return &OpenAIService{
		client: client,
	}
}

func (s *OpenAIService) GetAnalysis(ctx context.Context, prompt string) (string, error) {
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert data analyst specializing in tourism and accommodation industry analysis. Provide detailed, actionable insights based on the data provided.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   2000,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to get OpenAI response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
