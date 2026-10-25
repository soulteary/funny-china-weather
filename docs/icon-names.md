# Weather icon names (development version)

These changes are not part of the already published v0.2.0. The development
version uses canonical lowercase keys without `.svg` or a path. Old names
return an empty result / `false`; no compatibility aliases are provided.

## Naming rules

Use `condition[-intensity][-day|night]`, for example `rain-light` and
`snow-showers-night`. Names without a day/night suffix are shared across the
day. `variant-N` denotes alternate artwork, never an official weather intensity.
There are 34 resources: 23 automatically selected (including `unknown`) and
11 alternate/manual resources. Existing artwork is unchanged by the rename;
`unknown` is a new neutral question-mark SVG.

## Renamed keys

| Previous key | Canonical key |
| --- | --- |
| `day` | `clear-day` |
| `night` | `clear-night` |
| `cloudy` | `overcast` |
| `rainy-1` | `showers-day` |
| `rainy-night-1` | `showers-night` |
| `rainy-4` | `rain-light` |
| `rainy-5` | `rain-moderate` |
| `rainy-6` | `rain-heavy` |
| `thunder-rain` | `thunderstorm` |
| `icon-hail` | `thunderstorm-hail` |
| `snowy-day` | `snow-showers-day` |
| `snowy-night` | `snow-showers-night` |
| `snowy-4` | `snow-light` |
| `snowy-5` | `snow-moderate` |
| `snowy-6` | `snow-heavy` |
| `icon-sleet` | `freezing-rain` |
| `icon-fog` | `fog` |
| `icon-dust` | `dust` |
| `icon-wind` | `wind` |
| `thunder` | `lightning` |
| `cloudy-day-2` | `cloudy-day-variant-2` |
| `cloudy-day-3` | `cloudy-day-variant-3` |
| `cloudy-night-2` | `cloudy-night-variant-2` |
| `cloudy-night-3` | `cloudy-night-variant-3` |
| `rainy-2` | `showers-day-variant-2` |
| `snowy-2` | `snow-showers-day-variant-2` |
| `rainy-3` | `showers-day-variant-3` |
| `snowy-3` | `snow-showers-day-variant-3` |
| `rainy-7` | `rain-variant-7` |

Unchanged keys: `cloudy-day`, `cloudy-night`, `rain-snow-day`, `rain-snow-night`.
The new `unknown` key is the fallback for every unrecognized code.

## Automatic condition mapping

| Weather codes | Condition | Day key | Night key |
| --- | --- | --- | --- |
| 0 | 晴 | `clear-day` | `clear-night` |
| 1 | 多云 | `cloudy-day` | `cloudy-night` |
| 2 | 阴天 | `overcast` | `overcast` |
| 3 | 白天阵雨 | `showers-day` | `showers-night` |
| 4 | 雷阵雨 | `thunderstorm` | `thunderstorm` |
| 5 | 雷阵雨伴有冰雹 | `thunderstorm-hail` | `thunderstorm-hail` |
| 6 | 雨夹雪 | `rain-snow-day` | `rain-snow-night` |
| 7 | 小雨 | `rain-light` | `rain-light` |
| 8 | 中雨 | `rain-moderate` | `rain-moderate` |
| 9 | 大雨 | `rain-heavy` | `rain-heavy` |
| 10 | 暴雨 | `rain-heavy` | `rain-heavy` |
| 11 | 大暴雨 | `rain-heavy` | `rain-heavy` |
| 12 | 特大暴雨 | `rain-heavy` | `rain-heavy` |
| 13 | 阵雪 | `snow-showers-day` | `snow-showers-night` |
| 14 | 小雪 | `snow-light` | `snow-light` |
| 15 | 中雪 | `snow-moderate` | `snow-moderate` |
| 16 | 大雪 | `snow-heavy` | `snow-heavy` |
| 17 | 暴雪 | `snow-heavy` | `snow-heavy` |
| 18 | 有雾 | `fog` | `fog` |
| 19 | 冻雨 | `freezing-rain` | `freezing-rain` |
| 20 | 沙尘暴 | `dust` | `dust` |
| 21 | 小到中雨 | `rain-moderate` | `rain-moderate` |
| 22 | 中到大雨 | `rain-heavy` | `rain-heavy` |
| 23 | 大到暴雨 | `rain-heavy` | `rain-heavy` |
| 24 | 暴雨到大暴雨 | `rain-heavy` | `rain-heavy` |
| 25 | 大暴雨到特大暴雨 | `rain-heavy` | `rain-heavy` |
| 26 | 小到中雪 | `snow-moderate` | `snow-moderate` |
| 27 | 中到大雪 | `snow-heavy` | `snow-heavy` |
| 28 | 大到暴雪 | `snow-heavy` | `snow-heavy` |
| 29 | 浮尘 | `dust` | `dust` |
| 30 | 扬沙 | `dust` | `dust` |
| 31 | 强沙尘暴 | `dust` | `dust` |
| 32 | 浓雾 | `fog` | `fog` |
| 53 | 霾 | `fog` | `fog` |
| 301 | 雨 | `rain-light` | `rain-light` |
| 302 | 雪 | `snow-light` | `snow-light` |
| Other codes, including -1 | 未知 | `unknown` | `unknown` |

Daytime selection is 05:00 inclusive to 19:00 exclusive, in the supplied
time zone (`GetWeatherIconAt`) or process zone (`GetWeatherIconByCode`).
Only clear, cloudy, showers, rain/snow and snow showers have day/night pairs.

Several weather conditions intentionally share artwork: `rain-heavy` covers
multiple heavy-rain ranges; `fog` covers fog and haze; `dust` covers dust and
sandstorm conditions. Preserve the returned description or provider description
instead of inferring precise conditions from the resource name.
`freezing-rain` means code 19 (冻雨), separate from code 6 rain/snow (雨夹雪).

## Flare integration

Flare obtains `ConditionCode` from `GetWeatherIconByCode` and passes it to
`GetSVGCodeByName`. Both functions now agree on the new keys and keep their
existing signatures, including trusted `template.HTML` for inline SVG. Flare
does not need a call-site change or a mapping of its own.

Deploy/rebuild the application with one version of this library. Restart to
clear existing in-memory weather data. Any caller that hardcodes or persists
old keys must replace them using the mapping above; old serialized keys are
intentionally rejected. Public function compatibility does not mean resource
name compatibility. The adjacent Flare repository is not modified by this change.

## Validation

`go test ./...`, `go test -race ./...` and `go vet ./...` cover the canonical
resource inventory, rejection of removed keys, all selector-to-SVG pairs,
day/night boundaries, SVG composition/static output and Flare-style caching.
The test fixture `testdata/icon_names.json` records all renamed keys.
Historical v0.2.0 release notes retain the names shipped in that release.

Actual Flare validation was also run against an isolated copy of the adjacent
checkout with a local module replacement: its full `go test ./...` suite passed,
and `go test -race ./internal/settings/weather ./internal/pages/home` passed.
An additional 38-code check exercised Flare's actual `GetWeatherInfo`, the
current library's legacy SVG function and HTML template insertion with fixture
HTTP responses; it checked temperature, humidity, timestamp and unknown fallback.
No live provider or browser-animation result is implied by this offline check.

The reusable Flare-side test is saved at
`testdata/flare/canonical_weather_test.go`. To reproduce, copy Flare to a
temporary directory, put that test in its `internal/settings/weather/`
directory, and add a local `go.mod` replacement pointing this weather module to
the development checkout. Run the commands above from that temporary copy.
The fixture is under `testdata` so it does not add Flare dependencies to this
library or its normal tests.
