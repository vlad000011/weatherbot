package openweather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OpenWeatherClient struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) *OpenWeatherClient {
	return &OpenWeatherClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (o OpenWeatherClient) Coordinates(city string) (Coordinates, error) {

	url := "http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=5&appid=%s"
	resp, err := http.Get(fmt.Sprintf(url, city, o.apiKey))
	if err != nil {
		return Coordinates{}, err
	}
	if resp.StatusCode != 200 {
		return Coordinates{}, err
	}

	var coordinatesResponse []CoordinatesResponse
	err = json.NewDecoder(resp.Body).Decode(&coordinatesResponse)
	if err != nil {
		return Coordinates{}, err
	}

	if len(coordinatesResponse) == 0 {
		return Coordinates{}, fmt.Errorf("error empty")
	}
	return Coordinates{
		Lat: coordinatesResponse[0].Lat,
		Lon: coordinatesResponse[0].Lon,
	}, nil
}

func (o OpenWeatherClient) Weather(lat, lon float64) (Weather, error) {

	url := "https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric"
	resp, err := http.Get(fmt.Sprintf(url, lat, lon, o.apiKey))
	if err != nil {
		return Weather{}, fmt.Errorf("error %w", err)
	}
	if resp.StatusCode != 200 {
		return Weather{}, fmt.Errorf("error %d", resp.StatusCode)
	}
	var weatherResponse WeatherResponse
	err = json.NewDecoder(resp.Body).Decode(&weatherResponse)
	if err != nil {
		return Weather{}, fmt.Errorf("error %w", err)
	}
	return Weather{
		Temp: weatherResponse.Main.Temp, FeelsLike: weatherResponse.Main.FeelsLike, Humidity: weatherResponse.Main.Humidity,
		WindSpeed: weatherResponse.Wind.Speed, Description: weatherResponse.Weather[0].Description, MainWeather: weatherResponse.Weather[0].Main,
		CityName: weatherResponse.Name,
	}, nil
}
