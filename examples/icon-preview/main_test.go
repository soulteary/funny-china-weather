package main

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCompletePreview(t *testing.T) {
	icons, err := buildIcons("../../icons")
	if err != nil {
		t.Fatal(err)
	}
	if len(icons) != 49 {
		t.Fatalf("%d resources, want 49", len(icons))
	}
	automatic := 0
	for _, item := range icons {
		if item.Automatic {
			automatic++
		}
		checkPreviewCard(t, item)
	}
	if automatic != 38 {
		t.Fatalf("%d automatically matched, want 38", automatic)
	}
	checkPreviewPage(t)
}

func checkPreviewPage(t *testing.T) {
	t.Helper()
	page, err := render("../../icons")
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	if strings.Count(html, `<article class="card"`) != 49 || strings.Count(html, "<svg") != 98 {
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

func checkPreviewCard(t *testing.T, item icon) {
	t.Helper()
	if item.Description == "" || item.Codes == "" || !strings.Contains(string(item.Animated), "<svg") || !strings.Contains(string(item.Static), "<svg") {
		t.Errorf("incomplete card %s", item.Key)
	}
	checkPreviewMetadata(t, item)
	if strings.Contains(string(item.Static), "<animate") {
		t.Errorf("static %s retains SMIL", item.Key)
	}
}

func checkPreviewMetadata(t *testing.T, item icon) {
	t.Helper()
	if item.Key == "unknown" && (!item.Automatic || item.Description != "未知" || item.Animation != "静态图形") {
		t.Errorf("bad unknown metadata: %+v", item)
	}
	if item.Key == "freezing-rain" && (item.Description != "冻雨" || item.Codes != "19") {
		t.Errorf("bad freezing rain: %+v", item)
	}
	if item.Key == "wind" && item.Automatic {
		t.Fatal("wind should be manual")
	}
}

func TestPreviewInventoryFailures(t *testing.T) {
	for _, tc := range []struct{ name, filename, want string }{
		{"empty", "", "absent from directory"},
		{"unembedded", "missing.svg", "not embedded"},
		{"ignore non SVG", "README.md", "absent from directory"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.filename != "" {
				if err := os.WriteFile(filepath.Join(dir, tc.filename), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := buildIcons(dir)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
		})
	}
	if _, err := render(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing directory accepted")
	}
}

func TestPreviewMissingManualDescription(t *testing.T) {
	original := manualDescriptions
	manualDescriptions = map[string]string{}
	t.Cleanup(func() { manualDescriptions = original })
	if _, err := buildIcon("wind", automaticIconUsage()); err == nil || !strings.Contains(err.Error(), "missing description") {
		t.Fatalf("error=%v", err)
	}
}

func TestPreviewTemplateFailures(t *testing.T) {
	original := pageTemplate
	t.Cleanup(func() { pageTemplate = original })
	for _, source := range []string{"{{", "{{.Missing}}"} {
		pageTemplate = source
		if _, err := render("../../icons"); err == nil {
			t.Fatalf("invalid template %q accepted", source)
		}
	}
}

func TestMainWritesPreview(t *testing.T) {
	originalFlags, originalArgs := flag.CommandLine, os.Args
	t.Cleanup(func() { flag.CommandLine, os.Args = originalFlags, originalArgs })
	output := filepath.Join(t.TempDir(), "preview.html")
	flag.CommandLine = flag.NewFlagSet("icon-preview", flag.ContinueOnError)
	os.Args = []string{"icon-preview", "-icons", "../../icons", "-output", output}
	main()
	page, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(page), `<article class="card"`); got != 49 {
		t.Fatalf("written preview has %d cards, want 49", got)
	}
}

// Static SVG instance names intentionally vary; all other generated content
// must match the checked-in preview, including artwork and code labels.
func TestCheckedInPreviewIsCurrent(t *testing.T) {
	got, err := os.ReadFile("../../docs/icon-preview.html")
	if err != nil {
		t.Fatal(err)
	}
	want, err := render("../../icons")
	if err != nil {
		t.Fatal(err)
	}
	instance := regexp.MustCompile(`fcw-auto-[a-f0-9]{32}-[0-9]+`)
	if instance.ReplaceAllString(string(got), "INSTANCE") != instance.ReplaceAllString(string(want), "INSTANCE") {
		t.Fatal("preview is stale: run go run ./examples/icon-preview from the module root")
	}
}
