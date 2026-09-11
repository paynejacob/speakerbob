# speakerbob

A distributed soundboard: upload short audio clips, play them, and every
connected browser hears them via a live websocket broadcast. Go backend
(`pkg/`, `cmd/`) + a Vue 3/TypeScript frontend (`web/speakerbob/`) embedded
into the binary at build time.

## Layout

- `cmd/`, `main.go` — CLI entrypoint (cobra), server command, config
  loading.
- `pkg/` — application code. See `.claude/mappings/package-map.md` for
  what each package does.
- `web/speakerbob/` — frontend, out of scope for Go-focused work; see
  `.claude/rules/frontend/boundary.md`.
- `charts/`, `scripts/`, `Dockerfile`, `.github/workflows/` — build and
  release; see `.claude/mappings/build-release-map.md`.
- `docs/` — API specs (`openapi.yaml`, `asyncapi.yaml`) and
  `docs/superpowers/` design specs and plans for ongoing modernization
  work.

## Build / test

```bash
go build ./...
go test ./...
go generate ./pkg/...   # regenerate hotcereal provider code — see .claude/rules/golang/codegen.md
```

Before considering a Go change done, run the checks CI runs — see skill
`speakerbob-validate`.

## Rules

Note: each rule file's `paths:` frontmatter is documentation only (not
enforced by tooling) — every rule below loads into every session via the
imports regardless of what `paths:` lists.

@.claude/rules/common/coding-style.md
@.claude/rules/common/testing.md
@.claude/rules/golang/coding-style.md
@.claude/rules/golang/testing.md
@.claude/rules/golang/patterns.md
@.claude/rules/golang/codegen.md
@.claude/rules/frontend/boundary.md

## Mappings

@.claude/mappings/package-map.md
@.claude/mappings/request-flow.md
@.claude/mappings/build-release-map.md

## Learnings

@.claude/learnings/README.md
@.claude/learnings/2026-09-11-codegen-provider-files.md
@.claude/learnings/2026-09-11-frontend-out-of-scope.md

## Skills

Repo-specific skills are auto-discovered from `.claude/skills/`:
`speakerbob-codegen`, `speakerbob-validate`.
