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
		{"sunny", 0, "clear-day", "晴", "clear-night", "晴"},
		{"cloudy", 1, "cloudy-day", "多云", "cloudy-night", "夜间多云"},
		{"showers", 3, "showers-day", "白天阵雨", "showers-night", "夜间阵雨"},
		{"rain and snow", 6, "rain-snow-day", "雨夹雪", "rain-snow-night", "雨夹雪"},
		{"snow showers", 13, "snow-showers-day", "阵雪", "snow-showers-night", "阵雪"},
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
		{"thunderstorm includes rain", 4, "thunderstorm"},
		{"thunderstorm with hail highlights hail", 5, "thunderstorm-hail"},
		{"freezing rain uses frozen precipitation", 19, "freezing-rain"},
		{"light to moderate rain avoids the heaviest rain icon", 21, "rain-moderate"},
		{"light to moderate snow avoids the heaviest snow icon", 26, "snow-moderate"},
		{"haze has independent artwork", 53, "haze"},
		{"unspecified rain does not assume heavy intensity", 301, "rain-light"},
		{"unspecified snow does not assume heavy intensity", 302, "snow-light"},
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
	if key != "fog-dense" || name != "浓雾" {
		t.Fatalf("dense fog = (%q, %q), want dense fog artwork and the specific condition", key, name)
	}
}

func TestIconByCodeAtTimeIndependentCodes(t *testing.T) {
	codes := []int{2, 4, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 18, 19,
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 49, 53, 54, 55, 56, 57, 58, 301, 302,
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
	if utcKey != "showers-night" || chinaKey != "showers-day" {
		t.Fatalf("same instant should use supplied local hours: UTC %q, UTC+8 %q", utcKey, chinaKey)
	}
}

func TestGetWeatherIconByCodeUnknown(t *testing.T) {
	for _, code := range []int{-1, 123} {
		key, name := GetWeatherIconByCode(code)
		if key != "unknown" || name != "未知" {
			t.Errorf("GetWeatherIconByCode(%d) = (%q, %q), want unknown fallback", code, key, name)
		}
	}
}

func TestExtendedVisibilityCodes(t *testing.T) {
	body, err := os.ReadFile("testdata/weather_extensions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Source     string `json:"source"`
		Conditions []struct {
			Code        int    `json:"code"`
			Key         string `json:"key"`
			Description string `json:"description"`
		} `json:"conditions"`
	}
	if err := json.Unmarshal(body, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Source != "https://apidoc.weathercn.com/developers/best-practices.html" || len(fixture.Conditions) != 6 {
		t.Fatal("missing sourced extension cases")
	}
	for _, c := range fixture.Conditions {
		for _, hour := range []int{0, 12, 23} {
			key, name := GetWeatherIconAt(c.Code, time.Date(2026, 10, 26, hour, 0, 0, 0, time.UTC))
			if key != c.Key || name != c.Description {
				t.Errorf("code %d: got %s/%s, want %s/%s", c.Code, key, name, c.Key, c.Description)
			}
		}
	}
}

func TestWeatherIntensityArtworkIsDistinct(t *testing.T) {
	groups := [][]int{{18, 32, 49, 57, 58}, {53, 54, 55, 56}, {29, 30, 20, 31}, {9, 10, 11, 12}, {16, 17}}
	for _, codes := range groups {
		seen := map[string]bool{}
		for _, code := range codes {
			key, _ := GetWeatherIconAt(code, time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
			if seen[key] {
				t.Errorf("code %d shares intensity artwork %s", code, key)
			}
			seen[key] = true
			svg, ok := GetIconSVGWithPrefix(key, "compare")
			if !ok {
				t.Fatalf("missing %s", key)
			}
			for _, other := range codes {
				if other == code {
					continue
				}
				otherKey, _ := GetWeatherIconAt(other, time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
				otherSVG, _ := GetIconSVGWithPrefix(otherKey, "compare")
				if svg == otherSVG {
					t.Errorf("codes %d and %d have identical artwork", code, other)
				}
			}
		}
	}
	for code, want := range map[int]string{23: "rain-storm", 24: "rain-storm-severe", 25: "rain-storm-extreme", 28: "snow-storm"} {
		key, _ := GetWeatherIconAt(code, time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC))
		if key != want {
			t.Errorf("range code %d uses %s, want %s", code, key, want)
		}
	}
}

func TestDocumentedWeatherMappings(t *testing.T) {
	body, err := os.ReadFile("docs/icon-names.md")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[int]bool{}
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Split(line, "|")
		if len(fields) != 6 {
			continue
		}
		code, err := strconv.Atoi(strings.TrimSpace(fields[1]))
		if err != nil {
			continue
		}
		if seen[code] {
			t.Fatalf("duplicate documented code %d", code)
		}
		seen[code] = true
		for i, hour := range []int{12, 0} {
			key, name := GetWeatherIconAt(code, time.Date(2026, 10, 26, hour, 0, 0, 0, time.UTC))
			if key != strings.Trim(strings.TrimSpace(fields[3+i]), "`") {
				t.Errorf("code %d hour %d: documented key differs from %s", code, hour, key)
			}
			if i == 0 && name != strings.TrimSpace(fields[2]) {
				t.Errorf("code %d: documented label differs from %s", code, name)
			}
		}
	}
	if len(seen) != 42 {
		t.Errorf("documented %d codes, want 42", len(seen))
	}
	for code := range weatherIconsByCode {
		if !seen[code] {
			t.Errorf("code %d missing from documentation", code)
		}
	}
}

func TestSpecificIntensityMappings(t *testing.T) {
	for code, want := range map[int]string{10: "rain-storm", 11: "rain-storm-severe", 12: "rain-storm-extreme", 17: "snow-storm", 20: "sandstorm", 29: "dust", 30: "sand", 31: "sandstorm-strong", 32: "fog-dense", 53: "haze"} {
		for _, hour := range []int{0, 12} {
			key, _ := GetWeatherIconAt(code, time.Date(2026, 10, 26, hour, 0, 0, 0, time.UTC))
			if key != want {
				t.Errorf("code %d hour %d: %s, want %s", code, hour, key, want)
			}
		}
	}
}
