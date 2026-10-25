# Changelog

## Unreleased

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

Existing request functions keep their signatures and preserve raw timestamps.
Corrected condition descriptions may change display strings. Code that relied
on raw SVG bytes should use `GetIconSVGWithPrefix` with its own unique prefix.
