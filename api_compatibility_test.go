package weather_test

import (
	"context"
	"html/template"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

// Compile-time checks cover released compatibility and the extended public API.
var (
	_ func() (string, error)                                       = weather.GetMyIPLocation
	_ func(string) (int, int, int, string, error)                  = weather.GetWeatherByLocation
	_ func(int) (string, string)                                   = weather.GetWeatherIconByCode
	_ func(string) template.HTML                                   = weather.GetSVGCodeByName
	_ func(context.Context) (string, error)                        = weather.GetMyIPLocationContext
	_ func(context.Context, string) (int, int, int, string, error) = weather.GetWeatherByLocationContext
	_ func(context.Context, string) (weather.Observation, error)   = weather.GetObservation
	_ func(string) (time.Time, error)                              = weather.ParseUpdateTime
	_ func(int, time.Time) (string, string)                        = weather.GetWeatherIconAt
	_ func(string) (string, bool)                                  = weather.GetIconSVG
	_ func(string, string) (string, bool)                          = weather.GetIconSVGWithPrefix
	_ func(string) (string, bool)                                  = weather.GetStaticIconSVG
)
