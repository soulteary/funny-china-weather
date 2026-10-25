package weather_test

import (
	"html/template"

	weather "github.com/soulteary/funny-china-weather"
)

// Compile-time checks keep the original exported function signatures stable.
var (
	_ func() (string, error)                      = weather.GetMyIPLocation
	_ func(string) (int, int, int, string, error) = weather.GetWeatherByLocation
	_ func(int) (string, string)                  = weather.GetWeatherIconByCode
	_ func(string) template.HTML                  = weather.GetSVGCodeByName
)
