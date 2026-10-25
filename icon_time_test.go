package weather_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

func ExampleGetWeatherIconAt() {
	china := time.FixedZone("UTC+8", 8*60*60)
	at := time.Date(2026, time.October, 26, 8, 0, 0, 0, china)
	key, name := weather.GetWeatherIconAt(3, at)
	fmt.Println(key, name)
	// Output: showers-day 白天阵雨
}

func TestGetWeatherIconAtTimeZones(t *testing.T) {
	instant := time.Date(2026, time.October, 25, 23, 0, 0, 0, time.UTC)
	china := time.FixedZone("UTC+8", 8*60*60)
	for _, tc := range []struct {
		name string
		at   time.Time
		want string
	}{
		{"UTC night", instant, "clear-night"},
		{"China morning", instant.In(china), "clear-day"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if key, _ := weather.GetWeatherIconAt(0, tc.at); key != tc.want {
				t.Fatalf("icon = %q, want %q", key, tc.want)
			}
		})
	}
}

func TestIconVariantsReferenceExistingResources(t *testing.T) {
	for code := 0; code <= 32; code++ {
		for _, hour := range []int{8, 23} {
			at := time.Date(2026, time.October, 26, hour, 0, 0, 0, time.UTC)
			key, _ := weather.GetWeatherIconAt(code, at)
			if _, err := os.Stat("icons/" + key + ".svg"); err != nil {
				t.Errorf("code %d at hour %d refers to missing SVG %q: %v", code, hour, key, err)
			}
		}
	}
}
