# Changelog

## Unreleased

- Rename weather resource keys to explicit condition/intensity/day-night names;
  remove old-name lookup without aliases. See [the naming guide](docs/icon-names.md).
- Return a dedicated `unknown` icon for unrecognized weather codes, including at
  night, instead of sunny artwork. The embedded set now contains 34 SVG assets.
- Keep public function signatures and selector-to-SVG use compatible with Flare;
  literal old keys and persisted old-key caches must be migrated.

## v0.2.0

See the [release notes](docs/releases/v0.2.0.md) for migration from v0.1.1.

- Require Go 1.27.3 and use that version throughout local configuration and CI.
- Restore `GetSVGCodeByName(string) template.HTML` from the published v0.1.1
  API. It is deprecated in favor of `GetIconSVG`.
- Correct weather descriptions and icon selection against the China Weather
  legend, including code ranges, hail, freezing rain, haze, rain, and snow.
- Accept legitimate city names such as 津市市, normalize a small set of known
  prefecture and league aliases, and reject mismatched municipalities,
  malformed UTF-8, and invisible format characters.
- Correct SVG `viewBox` attributes and isolate IDs, classes, references, and
  animations for each returned instance. Add deterministic-prefix and static
  SVG APIs. The returned SVG bytes can change between calls.
- Add context-aware request APIs and an `Observation` API preserving the
  provider's descriptions. Add explicit UTC+8 timestamp and freshness helpers.
- Add structured provider errors while preserving wrapped cancellation and
  parsing causes, plus an exported response-size error.
- Return Tencent's actual `update_time` through the existing tuple API instead
  of the empty string returned by v0.1.1.
- Reject unsuccessful HTTP/business responses, incomplete observations, invalid
  numeric fields, and humidity outside 0–100. Bound JSON responses to 1 MiB and
  retain the three-second request timeout.
- Make icon selection deterministic, define daytime as 05:00 inclusive to
  19:00 exclusive, and use `day` / `未知` for unknown codes, including `-1`.
- Embed all 33 SVG resources with the standard library and remove the old
  minifier/parser module dependencies.
- Add a Flare-style consumer example with cache refresh, failure retry spacing,
  stale-data indicators and opt-in live cache/location integration tests.

Existing request functions keep their signatures and preserve raw timestamps.
Corrected condition descriptions may change display strings. Code that relied
on raw SVG bytes should use `GetIconSVGWithPrefix` with its own unique prefix.
The minimum Go version rises from v0.1.1's Go 1.17 to Go 1.27.3. Responses that
v0.1.1 accepted as partial or zero-valued success can now return errors; handle
errors before using weather data and use `errors.Is` / `errors.As` rather than
matching exact error text.
