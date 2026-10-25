package weather

import (
	"context"
	"errors"
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
	return getWeatherByLocation(context.Background(), location, defaultHTTPClient)
}

func getWeatherByLocation(ctx context.Context, location string, client httpDoer) (code int, degree int, humidity int, updateTime string, err error) {

	parsed, parseErr := parseLocation(location)
	if parseErr != nil {
		return 0, 0, 0, "", fmt.Errorf("tencent: invalid location: %w", parseErr)
	}
	query := url.Values{
		"source":       {"pc"},
		"weather_type": {"observe"},
		"province":     {parsed.province},
		"city":         {parsed.city},
	}

	response := new(responseTencentWeather)

	fetchErr := getJSONWithClient(ctx, client, "tencent", "https://wis.qq.com/weather/common?"+query.Encode(), response)
	if fetchErr != nil {
		return 0, 0, 0, "", fetchErr
	}

	if response.Status != 200 {
		return 0, 0, 0, "", fmt.Errorf("tencent: unexpected status %d: %s", response.Status, response.Message)
	}

	code, parseIntError := strconv.Atoi(response.Data.Observe.WeatherCode)
	if parseIntError != nil {
		return 0, 0, 0, "", fmt.Errorf("tencent: invalid weather_code: %w", parseIntError)
	}
	degree, parseIntError = strconv.Atoi(response.Data.Observe.Degree)
	if parseIntError != nil {
		return 0, 0, 0, "", fmt.Errorf("tencent: invalid degree: %w", parseIntError)
	}

	humidity, parseIntError = strconv.Atoi(response.Data.Observe.Humidity)
	if parseIntError != nil {
		return 0, 0, 0, "", fmt.Errorf("tencent: invalid humidity: %w", parseIntError)
	}

	if humidity < 0 || humidity > 100 {
		return 0, 0, 0, "", fmt.Errorf("tencent: humidity must be between 0 and 100; got %d", humidity)
	}

	updateTime = response.Data.Observe.UpdateTime
	if strings.TrimSpace(updateTime) == "" {
		return 0, 0, 0, "", errors.New("tencent: update_time is empty")
	}
	return code, degree, humidity, updateTime, nil
}
