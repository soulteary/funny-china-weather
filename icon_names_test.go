package weather

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCanonicalResourceNames(t *testing.T) {
	body, err := os.ReadFile("testdata/icon_names.json")
	if err != nil {
		t.Fatal(err)
	}
	var renamed map[string]string
	if err := json.Unmarshal(body, &renamed); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"cloudy-day": true, "cloudy-night": true, "rain-snow-day": true, "rain-snow-night": true, "unknown": true}
	for old, key := range renamed {
		if want[key] {
			t.Fatalf("duplicate canonical key %q", key)
		}
		want[key] = true
		if _, ok := GetIconSVG(old); ok {
			t.Errorf("removed key %q still accepted", old)
		}
		if _, ok := GetStaticIconSVG(old); ok {
			t.Errorf("removed static key %q still accepted", old)
		}
		if _, ok := GetIconSVGWithPrefix(old, "test"); ok {
			t.Errorf("removed prefixed key %q still accepted", old)
		}
		if GetSVGCodeByName(old) != "" {
			t.Errorf("removed legacy key %q still accepted", old)
		}
	}
	checkCanonicalAssets(t, want)
}

// Flare passes the selector's opaque key directly to the legacy template API.
// Check the complete mapping at both times, including extended and unknown codes.
func TestFlareSelectorToLegacySVG(t *testing.T) {
	codes := []int{-1, 53, 301, 302, 999}
	for code := 0; code <= 32; code++ {
		codes = append(codes, code)
	}
	for _, code := range codes {
		for _, hour := range []int{0, 12} {
			key, _ := GetWeatherIconAt(code, time.Date(2026, 10, 26, hour, 0, 0, 0, time.UTC))
			svg := GetSVGCodeByName(key)
			if !strings.Contains(string(svg), "<svg") {
				t.Errorf("Flare code %d hour %d key %q has no SVG", code, hour, key)
			}
		}
	}
	for _, hour := range []int{0, 12} {
		key, label := GetWeatherIconAt(-1, time.Date(2026, 10, 26, hour, 0, 0, 0, time.UTC))
		if key != "unknown" || label != "未知" {
			t.Fatalf("unknown fallback: %s %s", key, label)
		}
		unknown, _ := GetIconSVGWithPrefix(key, "compare")
		clear, _ := GetIconSVGWithPrefix("clear-day", "compare")
		if unknown == clear {
			t.Fatal("unknown weather reuses sunny artwork")
		}
	}
}

func checkCanonicalAssets(t *testing.T, want map[string]bool) {
	t.Helper()
	entries, err := iconFiles.ReadDir("icons")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(want) {
		t.Fatalf("%d assets, want %d", len(entries), len(want))
	}
	for _, entry := range entries {
		key := strings.TrimSuffix(entry.Name(), ".svg")
		if !want[key] {
			t.Errorf("unexpected resource %q", key)
		}
		if _, ok := GetIconSVG(key); !ok {
			t.Errorf("missing %q", key)
		}
		if _, ok := GetStaticIconSVG(key); !ok {
			t.Errorf("missing static %q", key)
		}
		if _, ok := GetIconSVGWithPrefix(key, "contract"); !ok {
			t.Errorf("missing fixed-prefix %q", key)
		}
		if GetSVGCodeByName(key) == "" {
			t.Errorf("legacy template API cannot read canonical key %q", key)
		}
	}
}
