# Releasing

The latest published tag may lag behind `main`. A merged pull request alone
does not update the version selected by plain `go get`. Tags through `v0.3.0`
already exist. Choose a new, unused version for the changes in `CHANGELOG.md` under `Unreleased`; do not reuse the historical
v0.3.0 release instructions as a plan for the next release.

Prepare version-specific release notes from the unreleased changes. The
[v0.3.0 notes](releases/v0.3.0.md) and
[v0.2.0 checklist](releases/v0.2.0-checklist.md) describe historical releases.

1. Merge the tested change and verify CI for the resulting `main` commit.
2. With Go 1.27.3, run `go test ./...`, `go test -race ./...`, and `go vet ./...`.
   Verify the published API compatibility tests as well as SVG composition,
   weather fixtures, location validation, and examples.
3. Move the intended `Unreleased` entries into the chosen version's changelog
   section. Confirm the README install command and the migration notes are included in the release commit. Merge the
   release-preparation pull request and confirm that CI passes for that exact
   resulting `main` commit. Do not tag an earlier feature-branch commit.
4. Create and push an annotated tag for the chosen version at that verified
   commit. Do not move or replace an existing tag. Include the changelog in the GitHub release notes.
5. In a fresh module, run the commands below. The proxy can take time to notice
   a new tag; confirm the actual selected version rather than assuming it has
   updated.

```sh
go mod init example.com/weather-release-check
# Set this to the new version selected above.
release_version=v0.X.Y
go get "github.com/soulteary/funny-china-weather@${release_version}"
go list -m github.com/soulteary/funny-china-weather
go list -m github.com/soulteary/funny-china-weather@latest
go doc github.com/soulteary/funny-china-weather.GetSVGCodeByName
go doc github.com/soulteary/funny-china-weather.GetObservation
go doc github.com/soulteary/funny-china-weather.GetIconSVGWithPrefix
```

Confirm both version queries select the intended new release,
then compile an external consumer of the legacy and new APIs. Preserve the
release's exact commit and validation results in its pull request or release
notes. Keep `go.mod`, `.go-version`, and the CI toolchain in agreement.

For v0.x releases, use the existing module path without a `/v0` suffix. A Go
library does not need compiled binaries attached to its release: the module
proxy downloads the tagged source, including embedded assets and license
notices. Verify the actual module ZIP and consumer behavior rather than only
the GitHub source archive.

Inspect the module ZIP against the release commit's resource inventory
(currently 49 canonical SVG files), both license notices, the complete gallery
and its generator. Verify removed keys are rejected, unknown weather returns `unknown`, and both README badge/report pairs
exist. The report-only bot commit has no new CI run because it uses `[skip ci]`;
verify its parent source commit has passed CI and that it changes only the five
stable report files before tagging it. Never tag an unchecked source change.
