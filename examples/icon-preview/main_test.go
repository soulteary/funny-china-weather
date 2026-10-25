package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestCompletePreview(t *testing.T) {
	icons, err := buildIcons("../../icons")
	if err != nil {
		t.Fatal(err)
	}
	if len(icons) != 34 {
		t.Fatalf("%d resources, want 34", len(icons))
	}
	automatic := 0
	for _, item := range icons {
		if item.Automatic {
			automatic++
		}
		if item.Description == "" || item.Codes == "" || !strings.Contains(string(item.Animated), "<svg") || !strings.Contains(string(item.Static), "<svg") {
			t.Errorf("incomplete card %s", item.Key)
		}
		if item.Key == "unknown" && (!item.Automatic || item.Description != "未知" || item.Animation != "静态图形") {
			t.Errorf("bad unknown metadata: %+v", item)
		}
		if item.Key == "freezing-rain" && (item.Description != "冻雨" || item.Codes != "19") {
			t.Errorf("bad freezing rain: %+v", item)
		}
		if item.Key == "wind" && item.Automatic {
			t.Fatal("wind should be manual")
		}
		if strings.Contains(string(item.Static), "<animate") {
			t.Errorf("static %s retains SMIL", item.Key)
		}
	}
	if automatic != 23 {
		t.Fatalf("%d automatically matched, want 23", automatic)
	}
	page, err := render("../../icons")
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	if strings.Count(html, `<article class="card"`) != 34 || strings.Count(html, "<svg") != 68 {
		t.Fatal("missing rendered cards or inline SVGs")
	}
	if strings.Contains(html, "&lt;svg") || strings.Contains(html, "ZgotmplZ") {
		t.Fatal("template escaped trusted SVG")
	}
	if strings.Contains(html, `<script src=`) || strings.Contains(html, `<link `) {
		t.Fatal("preview must work offline")
	}
	ids := map[string]bool{}
	for _, match := range regexp.MustCompile(`\bid="([^"]+)"`).FindAllStringSubmatch(html, -1) {
		if ids[match[1]] {
			t.Errorf("duplicate page ID %q", match[1])
		}
		ids[match[1]] = true
	}
}
