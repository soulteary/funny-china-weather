package weather_test

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	weather "github.com/soulteary/funny-china-weather"
)

func TestPublishedSVGAPI(t *testing.T) {
	tmpl := template.Must(template.New("icon").Parse(`<div>{{.}}</div>`))
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, weather.GetSVGCodeByName("day")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.String(), "<svg") || strings.Contains(rendered.String(), "&lt;svg") {
		t.Fatal("bundled SVG must remain usable as trusted html/template markup")
	}
	for _, key := range []string{"", "missing", "../day", "day.svg", `<script>alert(1)</script>`} {
		if got := weather.GetSVGCodeByName(key); got != "" {
			t.Errorf("unknown key %q returned markup", key)
		}
	}
}
