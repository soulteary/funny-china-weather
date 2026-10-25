package flare

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Exercise the real public API and its response validation without live services.
func TestProviderFailureRetryAndRecovery(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	calls := 0
	http.DefaultTransport = transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "wis.qq.com" || r.URL.Query().Get("province") != "北京市" {
			t.Fatalf("unexpected provider request: %s", r.URL)
		}
		body := `{"status":200,"data":{"observe":{"weather_code":"0","degree":"20","humidity":"50","update_time":"202610260800"}}}`
		if calls == 2 {
			body = `{"status":500,"message":"unavailable"}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	c := NewCache()
	at := time.Date(2026, 10, 26, 8, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	c.now = func() time.Time { return at }
	first := c.Get(context.Background(), "北京市")
	if !first.Available || first.Stale || first.Observation.TemperatureC != 20 || !strings.Contains(string(first.WeatherIcon), "<svg") {
		t.Fatalf("initial render: %+v", first)
	}
	at = at.Add(10*time.Minute - time.Nanosecond)
	if v := c.Get(context.Background(), "北京市"); v.Stale || calls != 1 || v.WeatherIcon == first.WeatherIcon {
		t.Fatal("cache must reuse observation but generate isolated SVGs")
	}
	at = at.Add(time.Nanosecond)
	failed := c.Get(context.Background(), "北京市")
	var providerError *weather.Error
	if !failed.Available || !failed.Stale || failed.Observation != first.Observation || !errors.As(failed.LastError, &providerError) || providerError.Stage != "business" {
		t.Fatalf("failed refresh: %+v", failed)
	}
	at = at.Add(time.Minute - time.Nanosecond)
	for i := 0; i < 20; i++ {
		c.Get(context.Background(), "北京市")
	}
	if calls != 2 {
		t.Fatalf("retry storm: %d requests", calls)
	}
	at = at.Add(time.Nanosecond)
	recovered := c.Get(context.Background(), "北京市")
	if calls != 3 || recovered.Stale || recovered.LastError != nil {
		t.Fatalf("recovery: %+v", recovered)
	}
}

func TestMissingDataLocationIsolationAndConcurrency(t *testing.T) {
	c := NewCache()
	at := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return at }
	calls := 0
	c.fetch = func(_ context.Context, location string) (weather.Observation, error) {
		calls++
		if location == "北京市" {
			return weather.Observation{UpdateTime: "202610260800"}, nil
		}
		return weather.Observation{}, context.DeadlineExceeded
	}
	c.Get(context.Background(), "北京市")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := c.Get(context.Background(), "上海市")
			if v.Available || !v.Stale || v.WeatherIcon != "" || !errors.Is(v.LastError, context.DeadlineExceeded) {
				t.Errorf("missing location must not show Beijing data: %+v", v)
			}
		}()
	}
	wg.Wait()
	if calls != 2 {
		t.Fatalf("duplicate concurrent refreshes: %d", calls)
	}
}

func TestObservationFreshness(t *testing.T) {
	for _, raw := range []string{"202610260700", "invalid", "202610260900"} {
		t.Run(raw, func(t *testing.T) {
			c := NewCache()
			c.now = func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) }
			c.fetch = func(context.Context, string) (weather.Observation, error) {
				return weather.Observation{UpdateTime: raw}, nil
			}
			if v := c.Get(context.Background(), "北京市"); !v.Available || !v.Stale {
				t.Fatalf("untrusted freshness: %+v", v)
			}
		})
	}
}

func ExampleCache_Get() {
	c := NewCache()
	c.now = func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) }
	c.fetch = func(context.Context, string) (weather.Observation, error) {
		return weather.Observation{TemperatureC: 20, HumidityPercent: 50, UpdateTime: "202610260800"}, nil
	}
	v := c.Get(context.Background(), "北京市")
	fmt.Printf("%s %d°C available=%t stale=%t\n", v.Location, v.Observation.TemperatureC, v.Available, v.Stale)
	// Output: 北京市 20°C available=true stale=false
}

func TestTemplateStatus(t *testing.T) {
	tmpl := template.Must(template.New("weather").Parse(`{{if .Available}}{{.WeatherIcon}} {{.Observation.TemperatureC}}°C{{if .Stale}} 天气已过期{{end}}{{else}}天气暂不可用{{end}}`))
	for _, tc := range []struct {
		name string
		view View
		want string
	}{
		{"fresh", View{Available: true, Observation: weather.Observation{TemperatureC: 20}}, " 20°C"},
		{"expired", View{Available: true, Stale: true, Observation: weather.Observation{TemperatureC: 20}}, " 20°C 天气已过期"},
		{"missing", View{Stale: true}, "天气暂不可用"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			if err := tmpl.Execute(&out, tc.view); err != nil {
				t.Fatal(err)
			}
			if out.String() != tc.want {
				t.Fatalf("rendered %q, want %q", out.String(), tc.want)
			}
		})
	}
}
