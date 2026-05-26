package openmeteo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GeocodingResponse — поиск города
type GeocodingResponse struct {
	Results []struct {
		ID        int     `json:"id"`
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

// WeatherResponse — данные погоды
type WeatherResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		FeelsLike   float64 `json:"apparent_temperature"`
		Humidity    int     `json:"relative_humidity_2m"`
		WindSpeed   float64 `json:"wind_speed_10m"`
		WeatherCode int     `json:"weather_code"`
	} `json:"current"`
}

type Weather struct {
	CityName    string
	Temp        float64
	FeelsLike   float64
	Humidity    int
	WindSpeed   float64
	Description string
}

func (c *Client) GetCoordinates(city string) (float64, float64, string, error) {
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=ru", city)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	var data GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, "", err
	}

	if len(data.Results) == 0 {
		return 0, 0, "", fmt.Errorf("city not found")
	}

	result := data.Results[0]
	return result.Latitude, result.Longitude, result.Name, nil
}

func (c *Client) GetWeather(lat, lon float64) (Weather, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.6f&longitude=%.6f&current=temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,weather_code&timezone=auto", lat, lon)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return Weather{}, err
	}
	defer resp.Body.Close()

	var data WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Weather{}, err
	}

	weather := Weather{
		Temp:      data.Current.Temperature,
		FeelsLike: data.Current.FeelsLike,
		Humidity:  data.Current.Humidity,
		WindSpeed: data.Current.WindSpeed,
	}

	// Простое описание погоды
	switch {
	case data.Current.WeatherCode == 0:
		weather.Description = "Ясно"
	case data.Current.WeatherCode >= 1 && data.Current.WeatherCode <= 3:
		weather.Description = "Облачно"
	case data.Current.WeatherCode >= 45 && data.Current.WeatherCode <= 48:
		weather.Description = "Туман"
	case data.Current.WeatherCode >= 51 && data.Current.WeatherCode <= 67:
		weather.Description = "Дождь"
	case data.Current.WeatherCode >= 71 && data.Current.WeatherCode <= 86:
		weather.Description = "Снег"
	default:
		weather.Description = "Переменная облачность"
	}

	return weather, nil
}
