package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"marketai/cards/internal/config"
	"marketai/cards/internal/domain"
	"net/http"
	"time"
)

type OpenAIService struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIService(cfg *config.Config) *OpenAIService {
	return &OpenAIService{
		apiKey: cfg.AI.OpenAIAPIKey,
		model:  cfg.AI.Model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *OpenAIService) GenerateCardContent(ctx context.Context, photoURL, description string) (*domain.GeneratedCard, error) {
	prompt := fmt.Sprintf(`
		Создай карточку товара для маркетплейса на основе описания: "%s"
		
		Требования:
		1. Заголовок должен быть кратким и привлекательным (до 60 символов)
		2. Описание должно быть подробным и продающим (150-300 слов)
		3. Теги должны быть релевантными для поиска (5-10 тегов)
		4. Используй эмодзи для привлекательности
		
		Ответь строго в формате JSON без пояснений, текста или Markdown.
		{
		  "title": "заголовок товара",
		  "description": "подробное описание товара",
		  "tags": ["тег1", "тег2", "тег3"]
		}
`, description)

	requestBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a helpful assistant.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream":      false,
		"max_tokens":  500,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	fmt.Println("resp===", resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Deepseek API error: status %d", resp.StatusCode)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var generatedCard domain.GeneratedCard
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &generatedCard); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Добавляем URL изображения (в реальном проекте здесь была бы генерация через DALL-E)
	generatedCard.Image = photoURL

	return &generatedCard, nil
}
