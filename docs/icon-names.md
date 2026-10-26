# Weather icon names

v0.3.0 replaces the keys shipped in v0.2.0 with canonical lowercase names
without `.svg` or a path. Old names
return an empty result / `false`; no compatibility aliases are provided.

## Naming rules

Use `condition[-intensity][-day|night]`, for example `rain-light` and
`snow-showers-night`. Names without a day/night suffix are shared across the
day. `variant-N` denotes alternate artwork, never an official weather intensity.
v0.4.0 adds 15 resources to the 34 shipped in v0.3.0.
There are 49 resources: 38 automatically selected (including `unknown`) and
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
| 10 | 暴雨 | `rain-storm` | `rain-storm` |
| 11 | 大暴雨 | `rain-storm-severe` | `rain-storm-severe` |
| 12 | 特大暴雨 | `rain-storm-extreme` | `rain-storm-extreme` |
| 13 | 阵雪 | `snow-showers-day` | `snow-showers-night` |
| 14 | 小雪 | `snow-light` | `snow-light` |
| 15 | 中雪 | `snow-moderate` | `snow-moderate` |
| 16 | 大雪 | `snow-heavy` | `snow-heavy` |
| 17 | 暴雪 | `snow-storm` | `snow-storm` |
| 18 | 有雾 | `fog` | `fog` |
| 19 | 冻雨 | `freezing-rain` | `freezing-rain` |
| 20 | 沙尘暴 | `sandstorm` | `sandstorm` |
| 21 | 小到中雨 | `rain-moderate` | `rain-moderate` |
| 22 | 中到大雨 | `rain-heavy` | `rain-heavy` |
| 23 | 大到暴雨 | `rain-storm` | `rain-storm` |
| 24 | 暴雨到大暴雨 | `rain-storm-severe` | `rain-storm-severe` |
| 25 | 大暴雨到特大暴雨 | `rain-storm-extreme` | `rain-storm-extreme` |
| 26 | 小到中雪 | `snow-moderate` | `snow-moderate` |
| 27 | 中到大雪 | `snow-heavy` | `snow-heavy` |
| 28 | 大到暴雪 | `snow-storm` | `snow-storm` |
| 29 | 浮尘 | `dust` | `dust` |
| 30 | 扬沙 | `sand` | `sand` |
| 31 | 强沙尘暴 | `sandstorm-strong` | `sandstorm-strong` |
| 32 | 浓雾 | `fog-dense` | `fog-dense` |
| 49 | 强浓雾 | `fog-dense-strong` | `fog-dense-strong` |
| 53 | 霾 | `haze` | `haze` |
| 54 | 中度霾 | `haze-moderate` | `haze-moderate` |
| 55 | 重度霾 | `haze-heavy` | `haze-heavy` |
| 56 | 严重霾 | `haze-severe` | `haze-severe` |
| 57 | 大雾 | `fog-heavy` | `fog-heavy` |
| 58 | 特强浓雾 | `fog-dense-extreme` | `fog-dense-extreme` |
| 301 | 雨 | `rain-light` | `rain-light` |
| 302 | 雪 | `snow-light` | `snow-light` |
| Other codes, including -1 | 未知 | `unknown` | `unknown` |

Daytime selection is 05:00 inclusive to 19:00 exclusive, in the supplied
time zone (`GetWeatherIconAt`) or process zone (`GetWeatherIconByCode`).
Only clear, cloudy, showers, rain/snow and snow showers have day/night pairs.

Fog, haze, blowing sand, sandstorms, rainstorm intensities and snowstorms
have independent artwork. Existing resource keys remain available; selector
results for these conditions now use the more specific keys above. Code 29
retains `dust` for suspended dust. Precipitation ranges use the upper intensity
(e.g. 23 uses `rain-storm`); generic 301/302 retain light rain/snow artwork.
Extensions 49 and 54–58 follow the [Huafeng WeatherCode table](https://apidoc.weathercn.com/developers/best-practices.html).
Resource availability does not establish live observation coverage.
Preserve descriptions to distinguish ranges sharing artwork.
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
name compatibility.
