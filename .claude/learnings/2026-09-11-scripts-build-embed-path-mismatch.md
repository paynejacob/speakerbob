---
id: scripts-build-embed-path-mismatch
trigger: "touching scripts/build, pkg/static/service.go, or release.yaml's cross-compiled binaries"
confidence: 0.9
domain: build
source: code-improvement
scope: project
---
# scripts/build embedded a stale frontend placeholder — RESOLVED

## Action

Historical context, kept for anyone wondering why `scripts/build` moves
its frontend build straight into `pkg/static/assets` rather than a
top-level `assets/` directory: `pkg/static/service.go`'s `//go:embed
assets` is package-relative (`pkg/static/assets/`), so that's the only
path that actually gets embedded. This was fixed — if you see a version
of `scripts/build` targeting a different path again, that's a
regression, not an intentional change.

## Evidence

- `pkg/static/service.go:16`: `//go:embed assets` (package-relative).
- Fixed in the commit "Fix scripts/build: wrong embed path, broken
  arm64 cross-compile" — `scripts/build` now does `rm -rf
  ../../pkg/static/assets && mv dist ../../pkg/static/assets` (relative
  to `web/speakerbob/`), matching what the `Dockerfile`'s `uibuild`
  stage already did correctly via a separate build path (`COPY
  --from=uibuild /ui/dist pkg/static/assets`).
- Before the fix: `pkg/static/assets/index.html` was a committed
  75-byte placeholder, and `scripts/build` moved its real build to
  repo-root `assets/` instead — a path nothing embeds.
