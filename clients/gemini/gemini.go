package gemini

import (
	"context"
	"fmt"

	"github.com/vlad000011/weatherbot/clients/openweather"
	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewClient(apiKey string, model string) (*GeminiClient, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать клиента: %w", err)

	}
	return &GeminiClient{
		client: client,
		model:  "gemini-3.5-flash",
	}, nil
}
func (c *GeminiClient) SuggestClothes(w openweather.Weather, city string) (string, error) {
	ctx := context.Background()
	prompt := fmt.Sprintf(`Ты — дружелюбный эксперт по выбору одежды.
Город: %s
Температура: %.1f°C
Ощущается как: %.1f°C
`,
		city,
		w.Temp,
		w.FeelsLike,
	)
	prompt += fmt.Sprintf(`Погода: %s
Влажность: %d%%
Ветер: %.1f м/с
Описание: %s

`,
		w.MainWeather,
		w.Humidity,
		w.WindSpeed,
		w.Description,
	)
	prompt += `Дай короткую рекомендацию, что человеку лучше надеть сегодня. 
Ответь на русском языке, в дружелюбном стиле, 2-4 предложения.`
	parts := []*genai.Part{
		{Text: prompt},
	}

	contents := []*genai.Content{
		{Parts: parts},
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("не удалось сгенерировать ответ: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini вернул пустой ответ")
	}
	return resp.Candidates[0].Content.Parts[0].Text, nil
}
