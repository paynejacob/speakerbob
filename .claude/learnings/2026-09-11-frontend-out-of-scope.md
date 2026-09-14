---
id: frontend-out-of-scope
trigger: "any task framed as Go version upgrade or backend code improvement"
confidence: 0.8
domain: frontend
source: repo-exploration
scope: project
---
# web/speakerbob is a separate app, out of scope for Go-focused work

## Action

Do not modify `web/speakerbob/` (Vue 3 + TypeScript frontend) code or
tests as part of Go version upgrade or backend code-quality work, unless
a task explicitly asks for frontend changes. See
`../rules/frontend/boundary.md`.

## Evidence

- User explicitly scoped the repo-modernization effort this way when
  asked whether the frontend was in scope: "include frontend for claude
  effort, but dont touch its code or tests."
- `web/speakerbob` has its own build tooling (yarn, `vue.config.js`,
  `babel.config.js`, `tsconfig.json`) entirely independent of the Go
  module — there's no shared build step other than `pkg/static` embedding
  its *build output*, not its source.
