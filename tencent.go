package weather

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type responseTencentWeather struct {
	Data struct {
		Observe struct {
			Degree        string `json:"degree"`
			Humidity      string `json:"humidity"`
			Precipitation string `json:"precipitation"`
			Pressure      string `json:"pressure"`
			UpdateTime    string `json:"update_time"`
			Weather       string `json:"weather"`
			WeatherCode   string `json:"weather_code"`
			WeatherShort  string `json:"weather_short"`
			WindDirection string `json:"wind_direction"`
			WindPower     string `json:"wind_power"`
		} `json:"observe"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// GetWeatherByLocation returns current weather and the unmodified upstream update time.
func GetWeatherByLocation(location string) (code int, degree int, humidity int, updateTime string, err error) {
	return GetWeatherByLocationContext(context.Background(), location)
}

// GetWeatherByLocationContext returns current weather while respecting ctx and the
// shared three-second client timeout. The upstream update time is returned unchanged.
// ctx must not be nil. An error returns zero weather values.
func GetWeatherByLocationContext(ctx context.Context, location string) (code int, degree int, humidity int, updateTime string, err error) {
	return getWeatherByLocation(ctx, location, defaultHTTPClient)
}

func getWeatherByLocation(ctx context.Context, location string, client httpDoer) (code int, degree int, humidity int, updateTime string, err error) {
	observation, err := getObservation(ctx, location, client)
	if err != nil {
		return 0, 0, 0, "", err
	}
	return observation.Code, observation.TemperatureC, observation.HumidityPercent, observation.UpdateTime, nil
}

func getObservation(ctx context.Context, location string, client httpDoer) (Observation, error) {
	parsed, parseErr := parseLocation(location)
	if parseErr != nil {
		return Observation{}, &Error{Provider: "tencent", Stage: "input", Field: "location", Err: fmt.Errorf("invalid location: %w", parseErr)}
	}
	query := url.Values{
		"source":       {"pc"},
		"weather_type": {"observe"},
		"province":     {parsed.province},
		"city":         {parsed.city},
	}
	response := new(responseTencentWeather)
	if err := getJSONWithClient(ctx, client, "tencent", "https://wis.qq.com/weather/common?"+query.Encode(), response); err != nil {
		return Observation{}, err
	}
	if response.Status != 200 {
		return Observation{}, &Error{Provider: "tencent", Stage: "business", BusinessCode: strconv.Itoa(response.Status), Err: fmt.Errorf("unexpected status %d: %s", response.Status, response.Message)}
	}
	code, err := strconv.Atoi(response.Data.Observe.WeatherCode)
	if err != nil {
		return Observation{}, &Error{Provider: "tencent", Stage: "field", Field: "weather_code", Err: fmt.Errorf("invalid weather_code: %w", err)}
	}
	degree, err := strconv.Atoi(response.Data.Observe.Degree)
	if err != nil {
		return Observation{}, &Error{Provider: "tencent", Stage: "field", Field: "degree", Err: fmt.Errorf("invalid degree: %w", err)}
	}
	humidity, err := strconv.Atoi(response.Data.Observe.Humidity)
	if err != nil {
		return Observation{}, &Error{Provider: "tencent", Stage: "field", Field: "humidity", Err: fmt.Errorf("invalid humidity: %w", err)}
	}
	if humidity < 0 || humidity > 100 {
		return Observation{}, &Error{Provider: "tencent", Stage: "field", Field: "humidity", Err: fmt.Errorf("humidity must be between 0 and 100; got %d", humidity)}
	}
	updated := response.Data.Observe.UpdateTime
	if strings.TrimSpace(updated) == "" {
		return Observation{}, &Error{Provider: "tencent", Stage: "field", Field: "update_time", Err: fmt.Errorf("update_time is empty")}
	}
	return Observation{
		Code:             code,
		TemperatureC:     degree,
		HumidityPercent:  humidity,
		Description:      response.Data.Observe.Weather,
		ShortDescription: response.Data.Observe.WeatherShort,
		UpdateTime:       updated,
	}, nil
}
