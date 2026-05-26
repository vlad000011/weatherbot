package gemini

import (
	"context"
	"fmt"

	"github.com/vlad000011/weatherbot/clients/weatherapi"
	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать клиента Gemini: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-3.5-flash",
	}, nil
}

func (c *GeminiClient) SuggestClothes(w weatherapi.Weather, city string) (string, error) {
	ctx := context.Background()

	prompt := fmt.Sprintf(`Ты — дружелюбный эксперт по одежде.
Город: %s
Температура: %.1f°C
Ощущается как: %.1f°C
Влажность: %d%%
Ветер: %.1f м/с
Описание: %s

Дай короткую рекомендацию, что надеть сегодня. Ответь на русском, 2-4 предложения.`,
		city,
		w.Temp,
		w.FeelsLike,
		w.Humidity,
		w.WindSpeed,
		w.Description,
	)

	parts := []*genai.Part{{Text: prompt}}
	contents := []*genai.Content{{Parts: parts}}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка Gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("пустой ответ")
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}
