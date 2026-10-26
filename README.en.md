# 晴雨笺 · funny-china-weather

A city's weather, at a glance.

[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md) [![Coverage](.github/coverage.svg)](.github/go-test-report.md)

[简体中文](README.md) | **English**

晴雨笺 (Qingyu Jian) is a lightweight Go library for current weather in China, outgoing-IP geolocation, and SVG icons selected by weather code. Its Chinese name evokes a small note about clear skies and rain. Use it in personal homepages, dashboards, and applications such as Flare.

- **49 embedded SVG icons**, with day/night variants, animation, and static output.
- Temperature, humidity, weather descriptions, and upstream update times, with request cancellation through `context`.
- Go standard library only. Weather and IP lookups need network access; icon selection and SVG lookup work offline.

## Icon preview

The animation below shows clear skies by day and night, clouds, showers, heavy snow, thunderstorms, fog, and wind. The GIF is approximately **252 KB**. After downloading the repository, open the [complete interactive gallery](docs/icon-preview.html) in a browser to search icons, switch light/dark canvases, or inspect static output.

[![Eight representative animated SVG weather icons](docs/icon-preview-animated.gif)](docs/icon-preview.html)

<details>
<summary>View the complete static gallery of all 49 icons</summary>

[![All 49 SVG weather icons, with names, conditions, and automatic matching status](docs/icon-preview.png)](docs/icon-preview.html)

</details>

## Installation and quick start

Requires **Go 1.27.3 or later**.

```sh
go get github.com/soulteary/funny-china-weather@main
```

These APIs are part of the current development version, so installation uses `@main`.

**Follow the location selection in [Tencent Weather](https://tianqi.qq.com/)**: search, select a specific city, district, or county, and retain its province/city/county fields. Check parent regions to distinguish places with the same name.

This example uses Tencent Weather's `湖南 → 常德 → 津市` selection to fetch weather and select an icon:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    weather "github.com/soulteary/funny-china-weather"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    selected := weather.Location{
        Province: "湖南",
        City:     "常德",
        County:   "津市",
    }
    observation, err := weather.GetObservationAt(ctx, selected)
    if err != nil {
        log.Fatal(err)
    }
    key, description := weather.GetWeatherIconByCode(observation.Code)
    fmt.Printf("%d°C, %d%% humidity, %s (%s), updated %s\n",
        observation.TemperatureC, observation.HumidityPercent,
        description, key, observation.UpdateTime)
}
```

`Observation` contains the weather code, temperature in degrees Celsius, relative humidity as an integer percentage, descriptions, and the original upstream update time. A typical timestamp is `202610250745` (`YYYYMMDDHHMM`), preserved as received. Unknown integer weather codes remain valid observation data and use the `unknown` icon.

When upgrading from v0.2.0, read the [icon naming and migration guide](docs/icon-names.md). v0.3.0 removes old icon names without aliases; calls that pass selector results directly to SVG lookup retain their signatures. Restart Flare after upgrading to clear cached old keys.

## Fetch weather

### Search and select a specific location

Follow the same flow as [Tencent Weather](https://tianqi.qq.com/): **enter a keyword → inspect candidates → confirm a specific place → fetch weather**. For example, `朝阳` can identify districts or cities in different regions; the keyword alone does not identify the intended place.

`SearchLocations` queries Tencent's location directory and returns `Location` candidates. Select and retain the intended candidate, then pass it to `GetObservationAt`. The library uses the provider's names without adding or removing administrative suffixes.

The following example searches for `津市`. A single candidate can be used directly; present multiple candidates for the user to choose:

```go
locations, err := weather.SearchLocations(ctx, "津市")
if err != nil {
    return err
}
if len(locations) != 1 {
    return fmt.Errorf("select one of %d locations", len(locations))
}
selected := locations[0]
observation, err := weather.GetObservationAt(ctx, selected)
if err != nil {
    return err
}
fmt.Println(selected.ID, selected.Province, selected.City, selected.County)
fmt.Println(observation.TemperatureC)
```

`Location` contains `ID` (weather city code), `Province`, `City`, and `County`. For example, `津市` resolves to `湖南 / 常德 / 津市`, producing `province=湖南&city=常德&county=津市`. Names retain Tencent's spelling. The code identifies the candidate; weather requests still use the location fields.

`SearchLocations` sorts candidates by code and returns an empty slice when none match. This order is not a recommendation ranking. Display province/city/county fields and use `ID` to identify a place. `GetObservationAt` fetches the selected location without searching again.

Search during configuration or user selection and retain `Location` for subsequent requests. Directory results are not cached automatically. Place the snippet in a function returning `error`, and import `fmt`.

### Context and structured observations

Set a deadline for an already selected `selected` location and check freshness according to your application policy. Place this snippet in a function returning `error`, and import `context`, `time`, and `fmt`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
observation, err := weather.GetObservationAt(ctx, selected)
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

`GetMyIPLocation` first gets the outgoing-IP city from IPIP, then searches Tencent with the unchanged city name and retains candidates whose province exactly matches IPIP's province. It does not rewrite province names or add suffixes. No match, multiple matches, or either provider failing returns an error. Directory errors identify `tencent`; IP lookup errors identify `ipip`.

IP geolocation is an optional way to choose a location automatically and reflects the outgoing connection, including proxies or VPNs. To specify a city, district, or county, search and select the intended place as in [Tencent Weather](https://tianqi.qq.com/).

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

`GetIconSVG` accepts an exact key such as `cloudy-day` and returns a complete SVG string. Paths and `.svg` extensions are rejected; invalid or missing keys return `"", false`. All 49 assets are embedded, so no extra asset generation or deployment is needed.

Each call isolates SVG IDs, references, CSS classes, and animation names with a unique instance prefix, allowing multiple icons in one HTML document. Source assets have valid `viewBox` attributes and scale with root dimensions. Separate calls for the same icon can produce different bytes.

- **Stable output:** use `GetIconSVGWithPrefix(key, prefix)`. Prefixes must start with an ASCII letter, contain only ASCII letters, digits, `_`, or `-`, and be at most 64 characters. Invalid prefixes return `"", false`.
- **Static output:** use `GetStaticIconSVG(key)` to remove SMIL animations and disable CSS animations. Applications can select it for users who prefer reduced motion.
- **Caching:** use a unique prefix for each DOM slot and include both the key and slot prefix in the cache key. Reusing one final SVG string in multiple positions duplicates IDs. Animated and static templates are compiled lazily per icon and reused; each render still applies its instance prefix. A fixed prefix provides stable output.

Weather descriptions follow the [China Weather legend](https://www.weather.com.cn/pub/legend.shtml), including extended codes `53` (haze), `301` (rain), and `302` (snow). Extensions `49` and `54–58` follow the [Huafeng WeatherCode table](https://apidoc.weathercn.com/developers/best-practices.html). Fog, haze, sandstorms, rainstorms and snowstorms have distinct artwork; precipitation ranges still share intensity icons. Keep the returned description rather than inferring precise conditions from the icon alone.

See the [naming guide](docs/icon-names.md) for all keys and code mappings. Regenerate the offline gallery with:

```sh
go run ./examples/icon-preview
```

See the [gallery instructions](examples/icon-preview/README.md) for controls and generation options.

## API and compatibility

Recommended APIs:

```go
func GetWeatherIconAt(code int, at time.Time) (key, name string)
func GetIconSVG(key string) (svg string, ok bool)
func GetIconSVGWithPrefix(key, prefix string) (svg string, ok bool)
func GetStaticIconSVG(key string) (svg string, ok bool)
func GetMyIPLocationContext(ctx context.Context) (string, error)
func SearchLocations(ctx context.Context, keyword string) ([]Location, error)
func GetObservationAt(ctx context.Context, location Location) (Observation, error)
func ParseUpdateTime(raw string) (time.Time, error)
```

`Observation` also provides `Age` and `IsStale`. See the [changelog](CHANGELOG.md) for version history and the [release guide](docs/releasing.md) for publication steps.

## Network behavior and errors

IP lookup calls IPIP's `https://myip.ipip.net/json`; location search calls Tencent's `https://i.news.qq.com/city/like`; weather lookup calls Tencent's `https://wis.qq.com/weather/common`. Availability, response formats, and city coverage depend on those providers.

Requests share an HTTP client with a **three-second timeout per HTTP request** and no automatic retries. Keyword weather lookup and IP location resolution each make two sequential requests; use a context deadline to limit the entire operation. Responses must have HTTP 2xx status, contain at most **1 MiB**, and decode as a single JSON value. Unknown fields are accepted.

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
