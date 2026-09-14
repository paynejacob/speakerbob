---
paths:
  - "web/**/*"
---
# Frontend Boundary

`web/speakerbob/` is a separate Vue 3 + TypeScript application (build
tooling: yarn, `vue.config.js`, `babel.config.js`, `tsconfig.json`). It is
documented here for context — see `../../mappings/build-release-map.md`
for how it fits into the Docker build — but it is **out of scope for code
changes** as part of Go-focused modernization work (Go version upgrade,
backend code/test improvements) unless a task explicitly asks for
frontend changes.

`pkg/static` embeds this app's build output via `//go:embed assets` in
`pkg/static/service.go` — a package-relative path, so it actually reads
from `pkg/static/assets/`. The `Dockerfile`'s `uibuild` stage populates
that directory correctly (`COPY --from=uibuild /ui/dist
pkg/static/assets`); `scripts/build` instead moves the frontend build to
repo-root `assets/`, which is not the embedded path — see
`../../mappings/build-release-map.md` and learning
`2026-09-11-scripts-build-embed-path-mismatch.md`. Changing how the
backend serves static assets is in scope; changing the frontend's own
source code is not.
