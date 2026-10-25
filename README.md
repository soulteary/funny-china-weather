# funny-china-weather

A small Go library for current weather in China and animated SVG weather icons.
It can locate the outgoing IP address, fetch Tencent observations for a province
and city, and select an embedded icon by weather code.

## Requirements and installation

Go 1.19 or later is required.

```sh
go get github.com/soulteary/funny-china-weather
```

Import the module with the package name `weather`:

```go
import weather "github.com/soulteary/funny-china-weather"
```

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
parse it or convert its time zone. Unknown integer weather codes remain valid
observation data.

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

The original public functions retain their signatures:

```go
func GetMyIPLocation() (string, error)
func GetWeatherByLocation(location string) (code int, degree int, humidity int, updateTime string, err error)
func GetWeatherIconByCode(weatherCode int) (string, string)
```

The additional functions are:

```go
func GetWeatherIconAt(code int, at time.Time) (key, name string)
func GetIconSVG(key string) (svg string, ok bool)
```

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
`errors.Is` and `errors.As` where applicable. The package currently exposes no
custom client, context, or retry configuration. Callers should handle errors
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

The GitHub Actions workflow tests Go 1.19.13 and Go 1.27.3 on Linux, checks the
race detector and `go vet`, and cross-compiles the library for Windows and macOS.
A manual workflow run can also enable the live integration checks. Regular
pushes and pull requests run the offline suite only.

## License and icon credits

The Go code is distributed under the [BSD 3-Clause License](LICENSE). The SVG
assets were selected and modified from
[Bram Kragten's weather-card](https://github.com/bramkragten/weather-card) and
[Bas Milius's weather-icons](https://github.com/basmilius/weather-icons).
Their MIT notices are retained in [icons/README.md](icons/README.md).
