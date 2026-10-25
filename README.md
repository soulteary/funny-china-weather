# funny-china-weather

A small Go library for current weather in China and animated SVG weather icons.
It can locate the outgoing IP address, fetch Tencent observations for a province
and city, and select an embedded icon by weather code.

## Requirements and installation

Go 1.27.3 or later is required. Development and CI use Go 1.27.3.

Install v0.2.0 after its version tag has been published:

```sh
go get github.com/soulteary/funny-china-weather@v0.2.0
```

Import the module with the package name `weather`:

```go
import weather "github.com/soulteary/funny-china-weather"
```

During release preparation, `@main` selects the current development code.
Merging changes does not publish a version: plain `go get` selects the latest
published tag. See the [v0.2.0 release checklist](docs/releases/v0.2.0-checklist.md)
for publication gates, the [release notes](docs/releases/v0.2.0.md) for migration
from v0.1.1, and [release steps](docs/releasing.md) for publishing and verifying
the version through the Go module proxy.

The library includes its SVG assets and has no runtime dependencies outside the
Go standard library. Location and observation requests require network access;
icon selection and SVG lookup work offline.

## Fetch current weather

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

`GetWeatherByLocation` returns the upstream weather code, temperature in degrees
Celsius, relative humidity as an integer percentage, and the original
`update_time` string. The provider currently supplies a timestamp such as
`202610250745` (`YYYYMMDDHHMM`). The library preserves it exactly and does not
parse it or convert its time zone automatically. Unknown integer weather codes
remain valid observation data.

For cancellation and the provider's original descriptions, use the structured
API:

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

Add `context` and `time` imports. `ParseUpdateTime` explicitly parses a strict
12-digit `YYYYMMDDHHMM` timestamp in UTC+8. `Observation.Age(at)` returns its age;
`IsStale(at, maxAge)` compares it with a caller-chosen positive limit. Parsing
rejects malformed timestamps and invalid calendar dates; `Age` and `IsStale`
also reject timestamps after the supplied reference time. Fetching an
observation still preserves any nonblank upstream timestamp. Age equal to the
limit is not stale. The library does not impose a default freshness policy.

The supported address forms include province-and-city names such as
`浙江省杭州市`, the four municipalities (`北京市`, `上海市`, `天津市`, and
`重庆市`), and autonomous regions with their city names. Examples for the five
autonomous regions are:

- `广西壮族自治区南宁市`
- `内蒙古自治区呼和浩特市`
- `宁夏回族自治区银川市`
- `西藏自治区拉萨市`
- `新疆维吾尔自治区乌鲁木齐市`

Surrounding whitespace and known short province or region names are normalized.
For example, `广西南宁市` resolves to `广西壮族自治区南宁市`, and `北京`
resolves to `北京市`. Legacy region names such as `广西省` are also accepted;
prefer the current official names in new code. A province must be followed by a
full city, autonomous-prefecture, prefecture, or league name ending in `市`,
`自治州`, `地区`, or `盟`. The municipality names stand alone.

District-only addresses, coordinates, and foreign locations are outside this
API's scope. Local address validation does not establish whether Tencent has
observations for every city.

Repeated administrative levels, mismatched municipality names, invalid UTF-8,
and invisible format characters are rejected. Legitimate names such as
`湖南省津市市` remain valid. The IP location normalizer explicitly recognizes
short forms for `黑龙江省大兴安岭地区`, `内蒙古自治区阿拉善盟`, and
`内蒙古自治区锡林郭勒盟`.
This is a small set of aliases, not a complete administrative-division database.

## Locate the outgoing IP address

```go
location, err := weather.GetMyIPLocation()
if err != nil {
    return err
}
code, degree, humidity, updated, err := weather.GetWeatherByLocation(location)
```

`GetMyIPLocation` looks up the IP address seen by IPIP. It reflects the outgoing
connection, including any proxy or VPN, rather than a device's GPS position.
The returned Chinese location uses normalized administrative names so it can be
passed directly to `GetWeatherByLocation`. Missing country, province, or city
information and locations outside China return an error.

`GetMyIPLocationContext(ctx)` and `GetWeatherByLocationContext(ctx, location)`
accept caller cancellation while returning the same values as their original
counterparts.

## Select an icon and read its SVG

```go
key, description := weather.GetWeatherIconByCode(3)
svg, ok := weather.GetIconSVG(key)
if !ok {
    return fmt.Errorf("missing weather icon %q", key)
}
fmt.Println(description, len(svg))
```

`GetWeatherIconByCode` uses the current local time of the process. Daytime starts
at **05:00 inclusive** and ends at **19:00 exclusive**. Sunny, cloudy, shower,
rain-and-snow, and snow-shower codes have separate day and night icons. Other
supported codes use the same icon throughout the day. Unknown codes, including
`-1`, return the key `day` and the description `未知`.

Use `GetWeatherIconAt` when a stored timestamp or a particular time zone should
determine the icon:

```go
chinaTime := time.FixedZone("UTC+8", 8*60*60)
at := time.Date(2026, time.October, 25, 19, 0, 0, 0, chinaTime)
key, description := weather.GetWeatherIconAt(3, at)
// key is "rainy-night-1"; description is "夜间阵雨".
```

Add `import "time"` for this example. `GetWeatherIconAt` uses the hour in the
supplied `time.Time` location and does not convert it to another time zone. The
upstream update string does not determine icon selection automatically.

`GetIconSVG` returns a complete SVG string for an exact icon key, such as
`cloudy-day`. It returns `"", false` for unknown keys, paths, or filenames with
an extension such as `cloudy-day.svg`. All 33 source SVG assets are embedded in
the library; no asset-generation step or external icon files are needed.

Every SVG returned by `GetIconSVG` receives a unique instance prefix for its
IDs, references, classes, and animations. Different icons and repeated copies
can be inserted inline into the same HTML document. The source assets have
valid `viewBox` attributes, so changing the root width and height scales their
content. Output bytes can differ between calls.

Use `GetIconSVGWithPrefix(key, prefix)` for stable output, for example when
caching or rendering on a server. Prefixes must begin with an ASCII letter,
contain only ASCII letters, digits, `_`, or `-`, and be at most 64 characters.
Use a unique prefix for every inline instance in a document. Invalid prefixes
return `"", false`. `GetStaticIconSVG(key)` provides an instance-isolated SVG
with SMIL and CSS animation disabled for applications that prefer reduced
motion. Applications should choose the static variant based on their own
accessibility preferences.

When caching inline SVGs, include both the icon key and a unique DOM-slot prefix
in the cache key. Copying one final SVG string into several positions duplicates
its IDs, even when that string originally came from `GetIconSVG`. Generate each
instance separately or use `GetIconSVGWithPrefix` with distinct slot prefixes.
Avoid depending on the bundled assets' internal IDs or CSS class names. A fixed
prefix makes output stable; it does not cache the parsing work automatically.

Weather descriptions follow the [China Weather legend](https://www.weather.com.cn/pub/legend.shtml),
including thunderstorm with hail, freezing rain, intensity ranges, and extended
codes `53` (haze), `301` (rain), and `302` (snow). An icon can approximate a
condition when the bundled set lacks an exact visual; the description retains
the condition's meaning.

The original public functions retain their signatures:

```go
func GetMyIPLocation() (string, error)
func GetWeatherByLocation(location string) (code int, degree int, humidity int, updateTime string, err error)
func GetWeatherIconByCode(weatherCode int) (string, string)
func GetSVGCodeByName(name string) template.HTML
```

The additional functions are:

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

`GetSVGCodeByName` is retained for callers of the published `v0.1.1` API and is
deprecated in favor of `GetIconSVG`. It returns only trusted bundled markup as
`html/template.HTML`, and returns an empty value for unknown names. Existing
function signatures are preserved; corrected weather descriptions and the
Go 1.27.3 minimum are intentional changes. See [the changelog](CHANGELOG.md).

## Network behavior and errors

The library calls IPIP's `https://myip.ipip.net/json` and Tencent's
`https://wis.qq.com/weather/common` observation endpoint. Their availability,
response format, and city coverage are controlled by those services. Successful
responses were checked on October 26, 2026, including the five autonomous-region
capital examples above. The endpoints currently require no API key.

Requests share a default HTTP client with a total timeout of three seconds per
request. The library does not retry automatically. Responses must have an HTTP
2xx status, contain at most 1 MiB, and decode as exactly one JSON value. Unknown
JSON fields are accepted.

IPIP must report `ret: "ok"`. Tencent must report `status: 200`; this business
status is checked even if an error response also contains observation fields.
The weather code, temperature, and humidity must parse as integers, humidity
must be between 0 and 100, and the update time must be nonblank. A blank or
malformed required field returns an error instead of partial weather data.

Errors identify the provider and failed request, response, or field. Underlying
request, read, JSON, and numeric conversion errors remain available through
`errors.Is` and `errors.As` where applicable. Use `errors.As(err, &providerError)`
with a `*weather.Error` to inspect `Provider`, `Stage`, `HTTPStatus`,
`BusinessCode`, and `Field`; HTTP 429 and provider business failures can be
handled separately. `errors.Is(err, context.Canceled)` and
`errors.Is(err, context.DeadlineExceeded)` preserve cancellation causes.
`ErrResponseTooLarge` identifies the response-size limit. The package exposes
no custom client or automatic retry configuration. Callers should handle errors
before using returned values.

## Development and validation

Normal tests use local fixtures and do not contact either provider:

```sh
go test ./...
go test -race ./...
go vet ./...
```

After modules are available locally, `GOPROXY=off go test ./...` checks that the
test suite runs without downloading dependencies. Examples that access the
providers are compiled during normal tests and are not executed automatically.
Deterministic icon examples run offline.

Live API tests are opt-in and can fail because of provider or network conditions:

```sh
go test -tags integration -run 'Integration$' -count=1 .
```

The GitHub Actions workflow uses Go 1.27.3 from `go.mod` on Linux, checks the
race detector and `go vet`, and cross-compiles the library for Windows and macOS.
A manual workflow run can also enable the live integration checks. Regular
pushes and pull requests run the offline suite only.

## License and icon credits

The Go code is distributed under the [BSD 3-Clause License](LICENSE). The SVG
assets were selected and modified from
[Bram Kragten's weather-card](https://github.com/bramkragten/weather-card) and
[Bas Milius's weather-icons](https://github.com/basmilius/weather-icons).
Their MIT notices are retained in [icons/README.md](icons/README.md).
