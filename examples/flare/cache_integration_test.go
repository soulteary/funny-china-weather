//go:build integration

package flare

import (
	"context"
	"html/template"
	"os"
	"strings"
	"testing"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

// Real provider requests with an advanced application clock: upstream timestamps
// remain real and need not change between consecutive requests.
func TestLiveCacheIntegration(t *testing.T) {
	c := NewCache()
	at := time.Now()
	c.now = func() time.Time { return at }
	calls := 0
	c.fetch = func(ctx context.Context, location string) (weather.Observation, error) {
		calls++
		return weather.GetObservation(ctx, location)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	first := c.Get(ctx, "北京市")
	if first.LastError != nil || !first.Available {
		t.Fatalf("initial live fetch: %+v", first)
	}
	if _, err := weather.ParseUpdateTime(first.Observation.UpdateTime); err != nil {
		t.Fatal(err)
	}
	cached := c.Get(ctx, "北京市")
	if calls != 1 || cached.Observation != first.Observation || cached.NextAttempt != first.NextAttempt {
		t.Fatal("cache hit changed data or fetched again")
	}
	at = first.NextAttempt.Add(-time.Nanosecond)
	c.Get(ctx, "北京市")
	if calls != 1 {
		t.Fatal("refreshed before expiry")
	}
	at = first.NextAttempt
	refreshed := c.Get(ctx, "北京市")
	if calls != 2 || refreshed.LastError != nil || !refreshed.Available || !refreshed.NextAttempt.Equal(at.Add(refreshInterval)) {
		t.Fatalf("live refresh: calls=%d view=%+v", calls, refreshed)
	}
	if first.WeatherIcon == cached.WeatherIcon || cached.WeatherIcon == refreshed.WeatherIcon || !strings.Contains(string(refreshed.WeatherIcon), "<svg") {
		t.Fatal("missing or reused inline SVG")
	}
	t.Logf("requests=%d; temperature=%d°C humidity=%d%% code=%d update=%s stale=%t; cache hit and expiry refresh passed", calls, refreshed.Observation.TemperatureC, refreshed.Observation.HumidityPercent, refreshed.Observation.Code, refreshed.Observation.UpdateTime, refreshed.Stale)
	writeLivePreview(t, []View{first, cached, refreshed})
}

func writeLivePreview(t *testing.T, views []View) {
	t.Helper()
	if path := os.Getenv("FLARE_LIVE_PREVIEW"); path != "" {
		page := template.Must(template.New("preview").Parse(`<!doctype html><html lang="zh"><meta charset="utf-8"><title>Flare live weather verification</title><style>body{font:18px system-ui;background:#f3f5f8;color:#202632;margin:48px}section{display:inline-block;background:white;padding:28px;margin:12px;border-radius:16px;vertical-align:top}svg{width:128px;height:128px}small{display:block;color:#667}</style><h1>Flare · 北京实时天气</h1>{{range .}}<section>{{.WeatherIcon}}<p>{{.Observation.TemperatureC}}°C · {{.Observation.HumidityPercent}}%</p><p>{{.ConditionText}}</p><small>{{.Observation.UpdateTime}}</small>{{if .Stale}}<p>天气已过期</p>{{end}}</section>{{end}}<p>首次获取 / 缓存命中 / 到期刷新</p></html>`))
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := page.Execute(f, views); err != nil {
			t.Fatal(err)
		}
	}
}
