# Flare-style consumer simulation

This package imports the current weather module directly. It models Flare's
home-page weather flow without changing the adjacent Flare repository.

Create one `cache := flare.NewCache()` for the application, then pass
`cache.Get(request.Context(), location)` to the page template on each request.
Follow [Tencent Weather](https://tianqi.qq.com/) to select a specific city,
district, or county. Search with `weather.SearchLocations` during configuration,
let the user choose the intended candidate, and pass its `Location.String()`
to the cache. This retains provider fields and skips directory lookup on refresh.
Do not construct an address by adding province or city suffixes.

For example, render the returned `View` with:

```html
{{if .Available}}
  <div class="weather">
    {{.WeatherIcon}}
    <span>{{.Observation.TemperatureC}}°C / {{.Observation.HumidityPercent}}%</span>
    <span>{{.ConditionText}}</span>
    {{if .Stale}}<span role="status">天气已过期</span>{{end}}
  </div>
{{else}}
  <span role="status">天气暂不可用</span>
{{end}}
```

The application policy is a ten-minute refresh after success and a one-minute
retry interval after failure, measured from request completion. Failures retain
the last successful data for that location and mark it stale. Missing data is
unavailable, never a fabricated zero-degree reading or another city's weather.
Provider timestamps older than thirty minutes, malformed timestamps and future
timestamps also mark data stale, even after a successful fetch. A successful
refresh clears the failure flag; timestamp freshness is still checked.

Icons are regenerated on each render to preserve unique inline SVG IDs. A mutex
serializes refreshes and prevents duplicate requests; this intentionally simple
example can block other readers during a refresh. It has no background refresh,
eviction or per-location concurrency, and is intended for a small configured
location set, like Flare's single-location home page.

Run `go test -race ./examples/flare` from the module root. Tests use a fake clock
and intercepted HTTP responses, exercise the real `GetObservation` validation,
and cover refresh/retry boundaries, retained data, recovery, timestamp freshness,
location isolation, concurrent requests and SVG instance isolation. They make
no live provider requests.

For an opt-in check against Tencent's real endpoint, run:

```sh
go test -tags integration -run TestLiveCacheIntegration -count=1 -v ./examples/flare
```

This makes two real requests, checks cache hits without extra requests, and
advances the application clock to the ten-minute expiry boundary to trigger
the second request. It does not wait ten wall-clock minutes or require upstream
data to change. Set `FLARE_LIVE_PREVIEW` to an HTML output path to save the
three inline SVG instances (first fetch, cache hit and refresh) for inspection.
Provider availability and freshness can vary; the test logs the stale flag
rather than assuming all successful responses contain fresh observations.
