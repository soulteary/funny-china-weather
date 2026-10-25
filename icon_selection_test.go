package weather

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestIconByCodeAtDayNightPairs(t *testing.T) {
	tests := []struct {
		name      string
		code      int
		dayKey    string
		dayName   string
		nightKey  string
		nightName string
	}{
		{"sunny", 0, "day", "晴", "night", "晴"},
		{"cloudy", 1, "cloudy-day", "多云", "cloudy-night", "夜间多云"},
		{"showers", 3, "rainy-1", "白天阵雨", "rainy-night-1", "夜间阵雨"},
		{"rain and snow", 6, "rain-snow-day", "雨夹雪", "rain-snow-night", "雨夹雪"},
		{"snow showers", 13, "snowy-day", "阵雪", "snowy-night", "阵雪"},
	}
	boundaries := []struct {
		name   string
		hour   int
		minute int
		second int
		day    bool
	}{
		{"midnight", 0, 0, 0, false},
		{"before sunrise", 4, 59, 59, false},
		{"sunrise", 5, 0, 0, true},
		{"noon", 12, 0, 0, true},
		{"before sunset", 18, 59, 59, true},
		{"sunset", 19, 0, 0, false},
		{"end of day", 23, 59, 59, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, boundary := range boundaries {
				t.Run(boundary.name, func(t *testing.T) {
					at := time.Date(2026, time.October, 25, boundary.hour, boundary.minute, boundary.second, 0, time.FixedZone("UTC+8", 8*60*60))
					key, name := iconByCodeAt(test.code, at)
					wantKey, wantName := test.nightKey, test.nightName
					if boundary.day {
						wantKey, wantName = test.dayKey, test.dayName
					}
					if key != wantKey || name != wantName {
						t.Fatalf("iconByCodeAt(%d, %s) = (%q, %q), want (%q, %q)", test.code, at.Format(time.RFC3339), key, name, wantKey, wantName)
					}
				})
			}
		})
	}
}

// This fixture records labels and image filenames from the official legend,
// independently of the implementation's icon and description table.
func TestWeatherDescriptionsMatchOfficialLegend(t *testing.T) {
	body, err := os.ReadFile("testdata/weather_legend.json")
	if err != nil {
		t.Fatal(err)
	}
	var legend struct {
		Source     string `json:"source"`
		Conditions []struct {
			Image       string `json:"image"`
			Description string `json:"description"`
		} `json:"conditions"`
	}
	if err := json.Unmarshal(body, &legend); err != nil {
		t.Fatal(err)
	}
	if legend.Source != "https://www.weather.com.cn/pub/legend.shtml" || len(legend.Conditions) == 0 {
		t.Fatal("official legend fixture is missing its source or entries")
	}
	for _, condition := range legend.Conditions {
		t.Run(condition.Image, func(t *testing.T) {
			if !strings.HasPrefix(condition.Image, "d") || !strings.HasSuffix(condition.Image, ".png") {
				t.Fatalf("unexpected official image filename %q", condition.Image)
			}
			code, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(condition.Image, "d"), ".png"))
			if err != nil {
				t.Fatal(err)
			}
			for _, hour := range []int{8, 23} {
				at := time.Date(2026, time.October, 26, hour, 0, 0, 0, time.UTC)
				key, name := GetWeatherIconAt(code, at)
				if name != condition.Description {
					t.Errorf("code %d at hour %d describes %q, official legend describes %q", code, hour, name, condition.Description)
				}
				if svg, ok := GetIconSVG(key); !ok || svg == "" {
					t.Errorf("official condition %q has no SVG for key %q", condition.Description, key)
				}
			}
		})
	}
}

func TestWeatherIconsHighlightDistinctConditions(t *testing.T) {
	cases := []struct {
		scenario string
		code     int
		key      string
	}{
		{"thunderstorm includes rain", 4, "thunder-rain"},
		{"thunderstorm with hail highlights hail", 5, "icon-hail"},
		{"freezing rain uses frozen precipitation", 19, "icon-sleet"},
		{"light to moderate rain avoids the heaviest rain icon", 21, "rainy-5"},
		{"light to moderate snow avoids the heaviest snow icon", 26, "snowy-5"},
		{"haze uses the available visibility icon", 53, "icon-fog"},
		{"unspecified rain does not assume heavy intensity", 301, "rainy-4"},
		{"unspecified snow does not assume heavy intensity", 302, "snowy-4"},
	}
	for _, test := range cases {
		t.Run(test.scenario, func(t *testing.T) {
			for _, hour := range []int{8, 23} {
				key, _ := GetWeatherIconAt(test.code, time.Date(2026, time.October, 26, hour, 0, 0, 0, time.UTC))
				if key != test.key {
					t.Errorf("code %d at hour %d uses %q, want %q", test.code, hour, key, test.key)
				}
			}
		})
	}
}

// The linked official legend does not list code 32, so its regression is separate.
func TestDenseFogRetainsItsSpecificDescription(t *testing.T) {
	key, name := GetWeatherIconAt(32, time.Date(2026, time.October, 26, 8, 0, 0, 0, time.UTC))
	if key != "icon-fog" || name != "浓雾" {
		t.Fatalf("dense fog = (%q, %q), want available fog artwork and the specific condition", key, name)
	}
}

func TestIconByCodeAtTimeIndependentCodes(t *testing.T) {
	codes := []int{2, 4, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 18, 19,
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 53, 301, 302,
		-999, -1, 33, 123}
	for _, code := range codes {
		at := time.Date(2026, time.October, 26, 0, 0, 0, 0, time.UTC)
		wantKey, wantName := GetWeatherIconAt(code, at)
		for _, hour := range []int{8, 12, 23} {
			key, name := GetWeatherIconAt(code, at.Add(time.Duration(hour)*time.Hour))
			if key != wantKey || name != wantName {
				t.Errorf("time-independent code %d changed at hour %d: (%q, %q) => (%q, %q)", code, hour, wantKey, wantName, key, name)
			}
		}
	}
}

func TestIconByCodeAtUsesSuppliedTimeZone(t *testing.T) {
	utc := time.Date(2026, time.October, 25, 2, 0, 0, 0, time.UTC)
	china := utc.In(time.FixedZone("UTC+8", 8*60*60))
	utcKey, _ := iconByCodeAt(3, utc)
	chinaKey, _ := iconByCodeAt(3, china)
	if utcKey != "rainy-night-1" || chinaKey != "rainy-1" {
		t.Fatalf("same instant should use supplied local hours: UTC %q, UTC+8 %q", utcKey, chinaKey)
	}
}

func TestGetWeatherIconByCodeUnknown(t *testing.T) {
	for _, code := range []int{-1, 123} {
		key, name := GetWeatherIconByCode(code)
		if key != "day" || name != "未知" {
			t.Errorf("GetWeatherIconByCode(%d) = (%q, %q), want unknown fallback", code, key, name)
		}
	}
}
