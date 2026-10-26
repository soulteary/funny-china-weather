package weather_test

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"testing"

	consumer "github.com/soulteary/flare/internal/settings/weather"
	weather "github.com/soulteary/funny-china-weather"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFlareCanonicalWeatherFlow(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	code := 0
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "wis.qq.com" {
			t.Fatalf("unexpected request: %s", r.URL)
		}
		body := fmt.Sprintf(`{"status":200,"data":{"observe":{"weather_code":"%d","degree":"20","humidity":"50","update_time":"202610260800"}}}`, code)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	tmpl := template.Must(template.New("weather").Parse(`<div class="pull-left">{{.WeatherIcon}}</div><div>{{.WeatherData.Degree}}°C {{.WeatherData.Humidity}}%</div>`))
	codes := []int{-1, 53, 301, 302, 999}
	for i := 0; i <= 32; i++ {
		codes = append(codes, i)
	}
	for _, value := range codes {
		code = value
		result, _, err := consumer.GetWeatherInfo("北京市")
		if err != nil {
			t.Fatalf("code %d: %v", code, err)
		}
		if result.Degree != 20 || result.Humidity != 50 || result.ExternalLastUpdate != "202610260800" {
			t.Fatalf("unexpected weather data: %+v", result)
		}
		svg := weather.GetSVGCodeByName(result.ConditionCode)
		if svg == "" {
			t.Fatalf("code %d returned missing resource %q", code, result.ConditionCode)
		}
		if (code == -1 || code == 999) && result.ConditionCode != "unknown" {
			t.Fatalf("unknown code selected %s", result.ConditionCode)
		}
		var output strings.Builder
		if err := tmpl.Execute(&output, map[string]any{"WeatherIcon": svg, "WeatherData": result}); err != nil {
			t.Fatal(err)
		}
		checkCanonicalTemplate(t, output.String(), code)
	}
	t.Logf("Flare's actual GetWeatherInfo and legacy SVG template flow passed for %d codes", len(codes))
}

func checkCanonicalTemplate(t *testing.T, output string, code int) {
	t.Helper()
	if !strings.Contains(output, "<svg") || strings.Contains(output, "&lt;svg") || !strings.Contains(output, "20°C 50%") {
		t.Fatalf("bad template output for code %d", code)
	}
}
