# Releasing

1. Choose a new, unused version and move the intended `Unreleased` entries in
   [CHANGELOG.md](../CHANGELOG.md) into that version's section. Update README
   installation commands and any migration notes for the release.
2. Keep `go.mod`, `.go-version` and CI on the same toolchain. Run
   `go test ./...`, `go test -race ./...` and `go vet ./...`.
3. Merge the release changes and verify CI for the exact resulting `main`
   commit. For a report-only `[skip ci]` commit, verify its source parent passed
   CI and that only generated report files changed.
4. Create and push an annotated tag at that verified commit, then publish the
   GitHub release with the version's changes and migration notes. Never move
   or replace an existing tag.
5. From a fresh module, install the explicit version and check `@latest`.
   Confirm the selected commit, compile a consumer of the public APIs, and
   inspect the module ZIP for the embedded SVGs, gallery and license notices.
   Module proxies may take time to discover a new tag.

Use the existing module path without a `/v0` suffix for v0.x releases. No
compiled binaries are required for this Go library.
