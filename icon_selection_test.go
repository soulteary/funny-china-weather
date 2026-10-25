package weather

import (
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
		{"snow showers", 13, "snowy-day", "雪天", "snowy-night", "雪天"},
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

func TestIconByCodeAtTimeIndependentCodes(t *testing.T) {
	tests := []struct {
		codes []int
		key   string
		name  string
	}{
		{[]int{2}, "cloudy", "阴天"},
		{[]int{4}, "thunder", "打雷"},
		{[]int{5}, "thunder-rain", "雷雨天"},
		{[]int{7}, "rainy-4", "小雨"},
		{[]int{8}, "rainy-5", "中雨"},
		{[]int{9}, "rainy-6", "大雨"},
		{[]int{10, 11, 12}, "rainy-6", "雨"},
		{[]int{14}, "snowy-4", "小雪"},
		{[]int{15}, "snowy-5", "中雪"},
		{[]int{16}, "snowy-6", "大雪"},
		{[]int{17}, "snowy-6", "暴雪"},
		{[]int{18, 32}, "icon-fog", "有雾"},
		{[]int{19, 21, 22, 23, 24, 25}, "rainy-6", "中雨"},
		{[]int{20, 29, 30, 31}, "icon-dust", "沙尘暴"},
		{[]int{26, 27, 28}, "snowy-6", "中雪"},
		{[]int{-999, -1, 33, 123}, "day", "未知"},
	}
	for _, test := range tests {
		for _, code := range test.codes {
			for _, hour := range []int{0, 12, 23} {
				at := time.Date(2026, time.October, 25, hour, 0, 0, 0, time.UTC)
				key, name := iconByCodeAt(code, at)
				if key != test.key || name != test.name {
					t.Errorf("iconByCodeAt(%d, hour %d) = (%q, %q), want (%q, %q)", code, hour, key, name, test.key, test.name)
				}
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
