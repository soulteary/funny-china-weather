package weather

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

func TestGetIconSVGAllResources(t *testing.T) {
	entries, err := iconFiles.ReadDir("icons")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 33 {
		t.Fatalf("embedded %d icons, want 33", len(entries))
	}
	for _, entry := range entries {
		entry := entry
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".svg") {
				t.Fatalf("unexpected embedded resource %q", entry.Name())
			}
			key := strings.TrimSuffix(entry.Name(), ".svg")
			svg, ok := GetIconSVG(key)
			if !ok || svg == "" {
				t.Fatalf("GetIconSVG(%q) = (%q, %v), want nonempty SVG and true", key, svg, ok)
			}

			decoder := xml.NewDecoder(strings.NewReader(svg))
			var root xml.Name
			var viewBox string
			for {
				token, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("invalid SVG XML: %v", err)
				}
				if start, ok := token.(xml.StartElement); ok && root.Local == "" {
					root = start.Name
					for _, attr := range start.Attr {
						if attr.Name.Local == "viewBox" {
							viewBox = attr.Value
						}
						if attr.Name.Local == "viewbox" {
							t.Fatal("incorrect lowercase viewbox attribute")
						}
					}
				}
			}
			if len(strings.Fields(viewBox)) != 4 {
				t.Fatalf("invalid root viewBox %q", viewBox)
			}
			if root.Local != "svg" || root.Space != "http://www.w3.org/2000/svg" {
				t.Fatalf("root element = %#v, want SVG namespace root", root)
			}
		})
	}
}

func TestGetIconSVGInvalidKeys(t *testing.T) {
	for _, key := range []string{
		"", "unknown", "README", "README.md", "Cloudy-day", "CLOUDY-DAY",
		"cloudy-day.svg", "cloudy-day.SVG", " cloudy-day", "cloudy-day ",
		"cloudy-day\n", ".", "..", "../day", "icons/day", "/day",
		"./day", `icons\day`, `..\day`, "day\x00", "晴", "%64ay", "day?",
	} {
		t.Run(key, func(t *testing.T) {
			svg, ok := GetIconSVG(key)
			if ok || svg != "" {
				t.Fatalf("GetIconSVG(%q) = (%q, %v), want empty SVG and false", key, svg, ok)
			}
		})
	}
}

func TestWeatherIconResourcesExist(t *testing.T) {
	// Cover all provider codes, special icons, and the unknown-code fallback.
	codes := []int{-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 123}
	for _, code := range codes {
		key, _ := GetWeatherIconByCode(code)
		if svg, ok := GetIconSVG(key); !ok || svg == "" {
			t.Errorf("weather code %d refers to missing icon %q", code, key)
		}
	}
}
