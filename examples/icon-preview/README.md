# Offline SVG gallery

From the module root, run:

```sh
go run ./examples/icon-preview
```

Open `docs/icon-preview.html` directly in a browser. The generated file contains
all 34 canonical SVG resources and needs no server, network access or external
assets. Each card labels the resource key, condition, weather codes, day/night
period, animation presence and automatic/manual matching status. The 23
automatic resources include the unknown-code fallback; 11 are manual variants.

Search by key, condition or code, filter by matching status, adjust the icon
size, or switch the icon canvas between light and dark. Pause controls both CSS
and supported SVG SMIL animations. The static switch shows the library's actual
`GetStaticIconSVG` output rather than merely pausing an animation. Reduced-motion
preferences pause animations initially. All SVG instance IDs are isolated,
including the hidden static versions, so the gallery also exercises inline
composition like Flare's templates.

Automatic weather labels are derived from the public selector at day/night
times. Manual variant descriptions describe artwork only and are not official
weather-code assignments. The source directory supplies the inventory, while
the actual SVG markup comes from this module's embedded-resource APIs.

Custom output and inventory paths:

```sh
go run ./examples/icon-preview -icons icons -output /tmp/weather-icons.html
```

After changing names, mappings or artwork, regenerate the checked-in HTML.
`go test ./examples/icon-preview` checks inventory, metadata, inline output,
static SVGs and page-wide ID uniqueness. Manual browser inspection is still
needed to assess animation timing and visual appearance.
