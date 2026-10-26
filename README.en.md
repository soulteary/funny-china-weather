# funny-china-weather

[简体中文](README.md) | **English**

[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md) [![Coverage](.github/coverage.svg)](.github/go-test-report.md)

A lightweight Go library for current weather in China, outgoing-IP geolocation, and SVG icons selected by weather code. Use it in personal homepages, dashboards, and applications such as Flare.

- **34 embedded SVG icons**, with day/night variants, animation, and static output.
- Temperature, humidity, weather descriptions, and upstream update times, with request cancellation through `context`.
- Go standard library only. Weather and IP lookups need network access; icon selection and SVG lookup work offline.

## Icon preview

The animation below shows clear skies by day and night, clouds, showers, heavy snow, thunderstorms, fog, and wind. The GIF is approximately **252 KB**. After downloading the repository, open the [complete interactive gallery](docs/icon-preview.html) in a browser to search icons, switch light/dark canvases, or inspect static output.

[![Eight representative animated SVG weather icons](docs/icon-preview-animated.gif)](docs/icon-preview.html)

<details>
<summary>View the complete static gallery of all 34 icons</summary>

[![All 34 SVG weather icons, with names, conditions, and automatic matching status](docs/icon-preview.png)](docs/icon-preview.html)

</details>

## Installation and quick start

Requires **Go 1.27.3 or later**.

```sh
go get github.com/soulteary/funny-china-weather@v0.3.0
```

Fetch Beijing's current weather and select its icon name:

```go
package main

import (
    "fmt"
    "log"

    weather "github.com/soulteary/funny-china-weather"
)

func main() {
    code, degree, humidity, updated, err := weather.GetWeatherByLocation("北京市")
    if err != nil {
        log.Fatal(err)
    }
    key, description := weather.GetWeatherIconByCode(code)
    fmt.Printf("%d°C, %d%% humidity, %s (%s), updated %s\n",
        degree, humidity, description, key, updated)
}
```

The return values are the weather code, temperature in degrees Celsius, relative humidity as an integer percentage, and the original upstream update time. A typical timestamp is `202610250745` (`YYYYMMDDHHMM`); fetching preserves it without parsing or converting its time zone. Unrecognized integer weather codes remain valid observation data and use the `unknown` icon.

When upgrading from v0.2.0, read the [icon naming and migration guide](docs/icon-names.md). v0.3.0 removes old icon names without aliases; calls that pass selector results directly to SVG lookup retain their signatures. Restart Flare after upgrading to clear cached old keys.

## Fetch weather

### Supported locations

Use municipality names directly: `北京市`, `上海市`, `天津市`, or `重庆市`. Elsewhere, provide a province or autonomous region followed by a city, for example:

- `浙江省杭州市`
- `广西壮族自治区南宁市`
- `内蒙古自治区呼和浩特市`
- `宁夏回族自治区银川市`
- `西藏自治区拉萨市`
- `新疆维吾尔自治区乌鲁木齐市`

Surrounding whitespace and known short names are normalized: `北京` → `北京市`, `广西南宁市` → `广西壮族自治区南宁市`. Historical forms such as `广西省` are also accepted; prefer current names in new code. A province must be followed by a full city, autonomous-prefecture, prefecture, or league name ending in `市`, `自治州`, `地区`, or `盟`.

District-only addresses, coordinates, and foreign locations are unsupported. Repeated administrative levels, mismatched municipality names, invalid UTF-8, and invisible format characters are rejected. Passing local validation does not establish whether Tencent has observations for that location.

Legitimate names such as `湖南省津市市` are accepted. IP geolocation also recognizes the known short forms `大兴安岭`, `阿拉善`, and `锡林郭勒`. These are a small alias set, not a complete administrative registry.

### Context and structured observations

Use `GetObservation` for cancellation and upstream weather descriptions. Place this snippet inside a function returning `error`, and add imports for `context`, `time`, and `fmt`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
observation, err := weather.GetObservation(ctx, "广西南宁市")
if err != nil {
    return err
}
fmt.Println(observation.Code, observation.TemperatureC,
    observation.HumidityPercent, observation.Description,
    observation.ShortDescription, observation.UpdateTime)
stale, err := observation.IsStale(time.Now(), 30*time.Minute)
if err != nil {
    return err
}
fmt.Println("stale:", stale)
```

`Description` and `ShortDescription` preserve upstream descriptions. Your application chooses the freshness policy:

- `ParseUpdateTime` strictly parses a 12-digit timestamp in UTC+8 and rejects invalid dates.
- `Observation.Age(at)` computes the age at a reference time; `IsStale(at, maxAge)` requires a positive maximum age.
- Observations after the reference time return an error; age equal to the maximum is not stale.
- Fetching only requires a nonblank update time. Call these methods explicitly to validate its format and freshness.

### Locate the outgoing IP address

```go
location, err := weather.GetMyIPLocation()
if err != nil {
    return err
}
code, degree, humidity, updated, err := weather.GetWeatherByLocation(location)
```

`GetMyIPLocation` returns a normalized Chinese location that can be passed directly to weather lookup. It reflects the outgoing connection, including proxies or VPNs. Missing country, province, or city data and locations outside China return an error.

Use `GetMyIPLocationContext(ctx)` and `GetWeatherByLocationContext(ctx, location)` for cancellation; their return values match the original functions.

## Use SVG icons

```go
key, description := weather.GetWeatherIconByCode(3)
svg, ok := weather.GetIconSVG(key)
if !ok {
    return fmt.Errorf("missing weather icon %q", key)
}
fmt.Println(description, len(svg))
```

`GetWeatherIconByCode` uses the current local time of the process. **Daytime is 05:00 inclusive to 19:00 exclusive.** Clear skies, clouds, showers, rain/snow, and snow showers have separate day/night icons; other conditions share an icon throughout the day. Unknown codes, including `-1`, return `unknown` and the description `未知`.

Use `GetWeatherIconAt` to specify a time and time zone, and import `time`:

```go
chinaTime := time.FixedZone("UTC+8", 8*60*60)
at := time.Date(2026, time.October, 25, 19, 0, 0, 0, chinaTime)
key, description := weather.GetWeatherIconAt(3, at)
// key is "showers-night"; description is "夜间阵雨".
```

It uses the hour in the supplied time's location without converting time zones. The upstream update time does not automatically determine icon selection.

### Resource names and inline SVG

`GetIconSVG` accepts an exact key such as `cloudy-day` and returns a complete SVG string. Paths and `.svg` extensions are rejected; invalid or missing keys return `"", false`. All 34 assets are embedded, so no extra asset generation or deployment is needed.

Each call isolates SVG IDs, references, CSS classes, and animation names with a unique instance prefix, allowing multiple icons in one HTML document. Source assets have valid `viewBox` attributes and scale with root dimensions. Separate calls for the same icon can produce different bytes.

- **Stable output:** use `GetIconSVGWithPrefix(key, prefix)`. Prefixes must start with an ASCII letter, contain only ASCII letters, digits, `_`, or `-`, and be at most 64 characters. Invalid prefixes return `"", false`.
- **Static output:** use `GetStaticIconSVG(key)` to remove SMIL animations and disable CSS animations. Applications can select it for users who prefer reduced motion.
- **Caching:** use a unique prefix for each DOM slot and include both the key and slot prefix in the cache key. Reusing one final SVG string in multiple positions duplicates IDs. A fixed prefix provides stable output; it does not automatically cache parsing.

Weather descriptions follow the [China Weather legend](https://www.weather.com.cn/pub/legend.shtml), including extended codes `53` (haze), `301` (rain), and `302` (snow). Some conditions share artwork, such as different rain intensities or fog and haze. Keep the returned description rather than inferring precise conditions from the icon alone.

See the [naming guide](docs/icon-names.md) for all keys and code mappings. Regenerate the offline gallery with:

```sh
go run ./examples/icon-preview
```

See the [gallery instructions](examples/icon-preview/README.md) for controls and generation options.

## API and compatibility

Original functions retained:

```go
func GetMyIPLocation() (string, error)
func GetWeatherByLocation(location string) (code int, degree int, humidity int, updateTime string, err error)
func GetWeatherIconByCode(weatherCode int) (string, string)
func GetSVGCodeByName(name string) template.HTML
```

Additional functions:

```go
func GetWeatherIconAt(code int, at time.Time) (key, name string)
func GetIconSVG(key string) (svg string, ok bool)
func GetIconSVGWithPrefix(key, prefix string) (svg string, ok bool)
func GetStaticIconSVG(key string) (svg string, ok bool)
func GetMyIPLocationContext(ctx context.Context) (string, error)
func GetWeatherByLocationContext(ctx context.Context, location string) (int, int, int, string, error)
func GetObservation(ctx context.Context, location string) (Observation, error)
func ParseUpdateTime(raw string) (time.Time, error)
```

`Observation` also provides `Age` and `IsStale`. `GetSVGCodeByName` is retained for v0.1.1 callers and returns trusted bundled SVG as `html/template.HTML`, or an empty value for unknown names. Prefer `GetIconSVG` in new code.

Function signature compatibility does not imply resource-name compatibility. See the [v0.3.0 release notes](docs/releases/v0.3.0.md) and [changelog](CHANGELOG.md) for migration. `@main` selects development code; merging does not publish a version. See the [release guide](docs/releasing.md) for publication steps.

## Network behavior and errors

IP lookup calls IPIP's `https://myip.ipip.net/json`; weather lookup calls Tencent's `https://wis.qq.com/weather/common`. Availability, response formats, and city coverage depend on those providers.

Requests share an HTTP client with a **three-second total timeout** and no automatic retries. Responses must have HTTP 2xx status, contain at most **1 MiB**, and decode as a single JSON value. Unknown fields are accepted.

IPIP must report `ret: "ok"`, and Tencent must report business `status: 200`. Weather code, temperature, and humidity must parse as integers; humidity must be between 0 and 100, and update time must be nonblank. Invalid required fields return an error rather than partial weather data.

Use `errors.As(err, &providerError)` with `*weather.Error` to inspect `Provider`, `Stage`, `HTTPStatus`, `BusinessCode`, and `Field`, distinguishing HTTP limits, business failures, and invalid fields. Underlying network, read, JSON, and numeric errors remain in the error chain. Use `errors.Is` to identify `context.Canceled`, `context.DeadlineExceeded`, and `weather.ErrResponseTooLarge`.

The library exposes no custom HTTP client or automatic retry configuration. Handle errors before using returned values.

## Development and validation

Normal tests use local fixtures without calling weather or IP services:

```sh
go test ./...
go test -race ./...
go vet ./...
```

With dependencies available locally, `GOPROXY=off go test ./...` checks offline testing. Live provider tests are opt-in and can fail because of network or upstream conditions:

```sh
go test -tags integration -run 'Integration$' -count=1 .
```

CI uses Go 1.27.3 on Linux, runs offline tests, race detection, and static checks, and cross-compiles for Windows and macOS. Manual workflow runs can enable live provider tests.

[Go Test Report Action](https://github.com/soulteary/go-test-report-action) and [Go Report Card Action](https://github.com/soulteary/goreportcard-action) generate coverage and quality reports for the library, Flare cache example, and SVG gallery example. Use the badges above to view reports. PR reports and raw results are workflow artifacts; CI updates report files after default-branch validation succeeds and the tested commit is still current. PR and tag builds never write reports back.

## License and icon credits

Go code uses the [BSD 3-Clause License](LICENSE). SVG assets were selected and modified from [Bram Kragten's weather-card](https://github.com/bramkragten/weather-card) and [Bas Milius's weather-icons](https://github.com/basmilius/weather-icons). MIT notices are retained in the [icon credits](icons/README.md).
