# Claude Project Scaffold — Design

## Context

This repo (`speakerbob`, a Go backend + Vue/TS frontend soundboard app) has no
`.claude/` folder or `CLAUDE.md` today. This is the first of three sequential
sub-projects for modernizing the repo:

- **A. Claude tooling scaffold** (this spec) — skills, rules, mappings,
  learnings under a project-level `.claude/` folder.
- **B. Go version upgrade** — `go.mod` currently pins Go 1.16; target a
  conservative recent version (1.22–1.23). CI (`.github/workflows/*.yaml`)
  and `Dockerfile` are also pinned to `1.16.6` and need updating.
- **C. Code/test-quality improvements** — incremental, scoped to code
  actually touched during B and follow-on work (not a full backfill).

Each sub-project gets its own design → plan → implementation cycle. This
spec covers **A only**.

## Repo facts this scaffold must capture

- Module: `github.com/paynejacob/speakerbob`, currently `go 1.16` in
  `go.mod`; local toolchain is 1.25.7.
- `pkg/*` packages: `auth` (incl. `auth/github`), `health`, `server`,
  `service`, `sound`, `static`, `store/badgerdb`, `version`, `websocket`.
- Code generation: `github.com/paynejacob/hotcereal` generates
  `zz_*_provider.go` files (e.g. `pkg/auth/zz_User_provider.go`,
  `pkg/sound/zz_Sound_provider.go`) via `go generate ./pkg/...`. These files
  must never be hand-edited.
- Test coverage today: only `pkg/sound/service_test.go` exists (1 of 34
  hand-written `.go` files).
- CI (`validate.yaml`) runs: `golang-lint` (go mod tidy + go fmt, diff-clean
  check), `golang-test` (`go test ./...`), `generate` (`go generate ./pkg/...`,
  diff-clean check), `web`, `helm`, `docker`.
- Frontend lives at `web/speakerbob/` (Vue 3 + TypeScript). It is documented
  for context but explicitly out of scope for code changes in this effort.
- The user's personal global `CLAUDE.md` already uses `@path` imports (e.g.
  `@RTK.md`) — this scaffold follows the same convention.
- The user has the `ecc` Claude plugin marketplace installed globally, which
  ships a `rules/common/` + `rules/<language>/` layered rules convention and
  a `continuous-learning-v2` atomic "instinct" format for learnings. This
  scaffold mirrors those *shapes* as static, git-committed files — it does
  **not** reimplement `ecc`'s hooks/observer/CLI infrastructure, since that
  already exists globally and is out of scope for a single project repo.

## Directory layout

```
CLAUDE.md                                  # root, short, @imports rules/mappings/learnings
.claude/
  rules/
    common/
      coding-style.md
      testing.md
    golang/
      coding-style.md                      # extends common/coding-style.md
      testing.md                           # extends common/testing.md
      patterns.md                          # hotcereal graph/store idioms used in this repo
      codegen.md                           # zz_*.go generation rule
    frontend/
      boundary.md                          # web/speakerbob is out of scope for code changes
  mappings/
    package-map.md                         # what each pkg/* package does, how they relate
    request-flow.md                        # HTTP + websocket request/broadcast flow
    build-release-map.md                   # scripts/, CI workflows, Dockerfile, charts/ tie-together
  learnings/
    README.md                              # explains the atomic-entry format, confidence scale
    2026-09-11-codegen-provider-files.md
    2026-09-11-frontend-out-of-scope.md
  skills/
    speakerbob-codegen/SKILL.md
    speakerbob-validate/SKILL.md
```

## Content design

### Rules

Each language/common rule file follows the `ecc` shape: YAML frontmatter with
a `paths:` glob list, an "> extends" reference line where applicable, and a
`## Reference` section pointing at relevant skills (either this repo's own,
e.g. `speakerbob-codegen`, or the globally-installed `ecc` skills like
`golang-testing`/`golang-patterns` when the content is generic Go advice
rather than repo-specific).

- `common/coding-style.md`, `common/testing.md`: universal, short —
  small interfaces, clear error propagation, tests should assert behavior
  not implementation.
- `golang/coding-style.md`: gofmt/goimports mandatory, wrap errors with
  `%w`, accept interfaces/return structs.
- `golang/testing.md`: table-driven `go test` with `testify` (already a
  dependency); HTTP-layer tests use `httpexpect` (already a dependency);
  storage-layer tests use an in-memory `badgerdb` instance rather than
  mocking the store; **states the "tests only for code we touch" policy**
  agreed for this effort.
- `golang/patterns.md`: how `hotcereal`'s `graph`/`store` packages are used
  in this repo (`auth`, `sound` domain types), what a provider is.
- `golang/codegen.md`: never hand-edit `zz_*.go`; after changing a graph
  type, run `go generate ./pkg/...`; run `scripts/validate/generate` to
  confirm the diff is clean, matching CI.
- `frontend/boundary.md`: `web/speakerbob` is a separate Vue/TS app; do not
  modify its code or tests as part of Go-focused work unless explicitly
  asked.

### Mappings

Factual reference docs, not instinct-formatted:

- `package-map.md`: one entry per `pkg/*` package — responsibility, key
  types, what it depends on.
- `request-flow.md`: how an HTTP request flows (`pkg/server` mux router →
  handler → `pkg/service`/domain service → `hotcereal` graph/store →
  `badgerdb`), and how the websocket broadcast path differs (`pkg/websocket`
  fan-out on sound play events).
- `build-release-map.md`: how `scripts/validate/*`, `.github/workflows/*`,
  `Dockerfile`, and `charts/` fit together for local validation, PR checks,
  and tagged releases.

### Learnings

Atomic entries mirroring the `continuous-learning-v2` instinct shape:

```yaml
---
id: <slug>
trigger: "<when this applies>"
confidence: 0.3-0.9
domain: <codegen|testing|build|go-version|frontend|...>
source: repo-exploration | go-upgrade | code-improvement
scope: project
---
# <Title>

## Action
<the rule/behavior to follow>

## Evidence
<what observation grounds this>
```

Seeded now with two entries (codegen files are generated;
frontend is a separate, out-of-scope app). Sub-projects B and C append new
entries as real gotchas surface (e.g. dependency behavior changes when
bumping Go version). `learnings/README.md` documents the format and
confidence scale so future entries stay consistent.

### Skills

Two curated, repo-specific skills (generic Go skills are covered by the
globally-installed `ecc` plugin, so they aren't duplicated here):

- `speakerbob-codegen`: walks through adding/modifying a `hotcereal`
  graph-backed type, running `go generate ./pkg/...`, and verifying the
  regenerated `zz_*.go` diff.
- `speakerbob-validate`: runs the same local checks CI runs before work is
  considered done — `go mod tidy` + `go fmt` diff-clean check, `go generate
  ./pkg/...` diff-clean check, `go test ./...`.

### Root `CLAUDE.md`

Short: one-paragraph repo overview, directory map, build/test commands,
`@.claude/rules/**/*.md`-style imports for rules/mappings/learnings (one
import line per file, mirroring the user's own global `CLAUDE.md` → `RTK.md`
pattern), and a pointer that skills are auto-discovered from
`.claude/skills/`.

## Validation

This scaffold is documentation/skills, not application code, so
"testing" means:

1. Every `SKILL.md` has valid YAML frontmatter (`name`, `description`) and
   non-empty body — mirrors `ecc`'s own `validate-skills.js` check.
2. Every `@import` path referenced in root `CLAUDE.md` resolves to a real
   file.
3. Each new skill is invoked once manually to confirm it loads and reads
   sensibly.
4. No existing file's behavior changes — this sub-project only adds new
   files.

## Non-goals

- No hook/observer/CLI infrastructure for learnings (that's `ecc`'s global
  system, already installed separately if the user wants it).
- No Go version changes, CI/Dockerfile edits, or application code changes
  (that's sub-project B).
- No test backfill for existing untested packages (that's sub-project C,
  and even then scoped to touched code only).
- No changes to `web/speakerbob`.
