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

`pkg/static` embeds this app's build output (`web/speakerbob`'s `dist/`,
copied to repo-root `assets/` by `scripts/build`, embedded via
`//go:embed assets` in `pkg/static/service.go`) and serves it as the
catch-all SPA fallback. Changing how the backend serves static assets is
in scope; changing the frontend's own source code is not.
