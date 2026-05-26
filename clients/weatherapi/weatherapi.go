package weatherapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

type WeatherResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`

	Current struct {
		TempC     float64 `json:"temp_c"`
		FeelsLike float64 `json:"feelslike_c"`
		Humidity  int     `json:"humidity"`
		WindKph   float64 `json:"wind_kph"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
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

func (c *Client) GetWeather(city string) (Weather, error) {
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&lang=ru", c.apiKey, city)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return Weather{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Weather{}, fmt.Errorf("город не найден")
	}

	var data WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Weather{}, err
	}

	return Weather{
		CityName:    data.Location.Name,
		Temp:        data.Current.TempC,
		FeelsLike:   data.Current.FeelsLike,
		Humidity:    data.Current.Humidity,
		WindSpeed:   data.Current.WindKph / 3.6, // km/h to m/s
		Description: data.Current.Condition.Text,
	}, nil
}
