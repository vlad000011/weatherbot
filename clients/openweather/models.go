package openweather

type CoordinatesResponse struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"` // ← Исправлено!
	Lon  float64 `json:"lon"`
}

type Coordinates struct {
	Lat float64
	Lon float64
}

type WeatherResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`

	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`

	Weather []struct {
		Description string `json:"description"`
		Main        string `json:"main"`
	} `json:"weather"`

	Name string `json:"name"` // Название города
}

// Удобная структура для использования в коде
type Weather struct {
	Temp        float64
	FeelsLike   float64
	Humidity    int
	WindSpeed   float64
	Description string
	MainWeather string
	CityName    string
}
