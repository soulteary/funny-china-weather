# Releasing

The latest published tag may lag behind `main`. A merged pull request alone
does not update the version selected by plain `go get`. The next release is
planned as `v0.2.0` because it raises the minimum Go version to 1.27.3 and
corrects weather descriptions while retaining existing function signatures.

1. Merge the tested change and verify CI for the resulting `main` commit.
2. With Go 1.27.3, run `go test ./...`, `go test -race ./...`, and `go vet ./...`.
   Verify the published API compatibility tests as well as SVG composition,
   weather fixtures, location validation, and examples.
3. Change the changelog's `Unreleased` heading to `v0.2.0`, replace the README
   development installation instructions with
   `go get github.com/soulteary/funny-china-weather@v0.2.0`, and commit those
   release-documentation updates. Confirm that CI passes for that exact commit.
4. Create and push an annotated `v0.2.0` tag at that verified commit. Do not move
   or replace an existing tag. Include the changelog in the GitHub release notes.
5. In a fresh module, run the commands below. The proxy can take time to notice
   a new tag; confirm the actual selected version rather than assuming it has
   updated.

```sh
go mod init example.com/weather-release-check
go get github.com/soulteary/funny-china-weather@v0.2.0
go list -m github.com/soulteary/funny-china-weather
go list -m github.com/soulteary/funny-china-weather@latest
go doc github.com/soulteary/funny-china-weather.GetSVGCodeByName
go doc github.com/soulteary/funny-china-weather.GetObservation
go doc github.com/soulteary/funny-china-weather.GetIconSVGWithPrefix
```

Confirm both version queries select `v0.2.0` (or a newer intended release),
then compile an external consumer of the legacy and new APIs. Preserve the
release's exact commit and validation results in its pull request or release
notes. Keep `go.mod`, `.go-version`, and the CI toolchain in agreement.
