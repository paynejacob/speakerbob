# Claude Project Scaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the project-level `.claude/` scaffold (rules, mappings, learnings, skills) plus a root `CLAUDE.md` for the speakerbob repo — sub-project A of the repo modernization effort.

**Architecture:** Pure-documentation/skills scaffold, no application code changes. Root `CLAUDE.md` stays short and `@`-imports individual rule/mapping/learning files (mirroring the user's own global `CLAUDE.md` → `RTK.md` pattern). Rules follow the globally-installed `ecc` plugin's layered `common/` + language-layer shape. Learnings follow `ecc`'s `continuous-learning-v2` atomic "instinct" format, as static git-committed files (no hooks/observer/CLI infra — that's out of scope). Skills are curated, repo-specific only.

**Tech Stack:** Markdown + YAML frontmatter only. No Go code, no build tooling changes.

**Spec:** `docs/superpowers/specs/2026-09-11-claude-scaffold-design.md`

## Global Constraints

- No Go version, CI workflow, Dockerfile, or application code changes in this plan — that's sub-project B (Go upgrade), tracked separately.
- No test backfill for existing untested packages — that's sub-project C, and even then scoped to touched code only.
- No changes to `web/speakerbob/` (frontend) — documented, not modified.
- Every `SKILL.md` must have valid YAML frontmatter with non-empty `name` (matching its directory name) and `description`, and a non-empty body.
- Every `@`-import path referenced in root `CLAUDE.md` must resolve to a real file in the repo.
- Rule files use YAML frontmatter with a `paths:` glob list; language-specific rule files include a `> This file extends [common/x.md](../common/x.md)...` line.
- Learning files use the atomic-instinct frontmatter shape: `id`, `trigger`, `confidence`, `domain`, `source`, `scope`, followed by `## Action` and `## Evidence` sections.
- This plan only adds new files — no existing file's content or behavior changes.

---

## File Structure

```
CLAUDE.md                                          # new, repo root
.claude/
  rules/
    common/coding-style.md                         # new
    common/testing.md                               # new
    golang/coding-style.md                          # new
    golang/testing.md                                # new
    golang/patterns.md                                # new
    golang/codegen.md                                 # new
    frontend/boundary.md                              # new
  mappings/
    package-map.md                                    # new
    request-flow.md                                    # new
    build-release-map.md                                # new
  learnings/
    README.md                                           # new
    2026-09-11-codegen-provider-files.md                # new
    2026-09-11-frontend-out-of-scope.md                 # new
  skills/
    speakerbob-codegen/SKILL.md                         # new
    speakerbob-validate/SKILL.md                        # new
```

---

### Task 1: Rules — common + golang layers

**Files:**
- Create: `.claude/rules/common/coding-style.md`
- Create: `.claude/rules/common/testing.md`
- Create: `.claude/rules/golang/coding-style.md`
- Create: `.claude/rules/golang/testing.md`
- Create: `.claude/rules/golang/patterns.md`
- Create: `.claude/rules/golang/codegen.md`

**Interfaces:**
- Consumes: N/A (standalone documentation).
- Produces: file paths `.claude/rules/common/coding-style.md`, `.claude/rules/common/testing.md`, `.claude/rules/golang/coding-style.md`, `.claude/rules/golang/testing.md`, `.claude/rules/golang/patterns.md`, `.claude/rules/golang/codegen.md` — referenced by Task 7's root `CLAUDE.md` imports.

- [ ] **Step 1: Create `.claude/rules/common/coding-style.md`**

```markdown
---
paths:
  - "**/*"
---
# Coding Style (Common)

Universal principles that apply regardless of language.

## Interfaces and boundaries

- Keep interfaces small — one to three methods. A caller should be able to
  understand what a dependency does without reading its implementation.
- Prefer explicit error returns over exceptions/panics for expected failure
  modes. Reserve panics for programmer errors (invariant violations), never
  for expected runtime conditions like "not found" or "invalid input".

## Naming

- Names should say what a thing is or does, not how it's implemented.
- Avoid abbreviations that aren't immediately obvious to a new reader.

## Comments

- Default to no comments. Only add one when the *why* is non-obvious: a
  hidden constraint, a workaround for a specific bug, a subtle invariant.
- Never restate what the code already says through naming.

## Language note

Language-specific rule files (`../golang/`, etc.) may override any of the
above where the language's idioms differ.
```

- [ ] **Step 2: Create `.claude/rules/common/testing.md`**

```markdown
---
paths:
  - "**/*"
---
# Testing (Common)

## Scope for this repo

Only 1 of 34 hand-written Go files in this repo has a test today
(`pkg/sound/service_test.go`). The agreed policy for this modernization
effort: **add or extend tests only for code that a change actually
touches** — this is not a full test-coverage backfill project. See
`../golang/testing.md` for Go-specific mechanics.

## Principles

- Test behavior, not implementation. A test should still pass after an
  internal refactor that doesn't change observable behavior.
- A failing test should tell you what's wrong without needing to read the
  implementation — assert on specific expected values, not just "no error".
- Prefer one focused assertion path per test case over one giant test that
  exercises many unrelated behaviors.
```

- [ ] **Step 3: Create `.claude/rules/golang/coding-style.md`**

```markdown
---
paths:
  - "**/*.go"
  - "**/go.mod"
  - "**/go.sum"
---
# Go Coding Style

> This file extends [common/coding-style.md](../common/coding-style.md) with Go specific content.

## Formatting

- `gofmt` is mandatory — CI (`scripts/validate/golang-lint`) fails the build
  on any diff after `go mod tidy && go fmt ./...`. Run both locally before
  committing.

## Design principles

- Accept interfaces, return concrete structs.
- Wrap errors with context using `%w`, never swallow them:

  ```go
  if err != nil {
      return fmt.Errorf("failed to save sound %q: %w", sound.Id, err)
  }
  ```

- `pkg/service.SpeakerbobError` / `NotAcceptableError` are this repo's
  existing error-to-HTTP-response types (`pkg/service/errors.go`). Use
  `service.WriteErrorResponse` when writing new HTTP handlers instead of
  inventing a new error-response shape.

## Reference

See the globally-installed `ecc` plugin skill `golang-patterns` for
generic Go idioms not specific to this repo. `golang/patterns.md` in this
folder covers idioms specific to *this* codebase (hotcereal graph/store).
```

- [ ] **Step 4: Create `.claude/rules/golang/testing.md`**

```markdown
---
paths:
  - "**/*.go"
---
# Go Testing

> This file extends [common/testing.md](../common/testing.md) with Go specific content.

## Framework and libraries already in this repo

- Standard `go test` with table-driven tests.
- `github.com/stretchr/testify` (already a dependency) for assertions —
  `assert`/`require`.
- `github.com/gavv/httpexpect/v2` (already a dependency) for HTTP handler
  tests — wrap the `*mux.Router` a service registers its routes on.
- For storage-layer tests, use a real in-memory-mode `badgerdb.Store`
  (badger supports `badger.DefaultOptions("").WithInMemory(true)`) rather
  than mocking `hotcereal/pkg/store.Store` — the existing
  `pkg/sound/service_test.go` is the reference example for this pattern.

## Running tests

```bash
go test ./...
go test -race ./...
go test -cover ./...
```

CI (`scripts/validate/golang-test`) runs plain `go test ./...` (no `-race`
flag currently) — match that as the minimum bar, `-race` locally is a good
extra check when touching `pkg/websocket` or `pkg/sound/queue.go`
(concurrent code).

## Coverage policy

Tests only for code a change touches (see `../common/testing.md`) — do not
open a separate PR-shaped effort to backfill coverage on untouched
packages.
```

- [ ] **Step 5: Create `.claude/rules/golang/patterns.md`**

```markdown
---
paths:
  - "pkg/**/*.go"
---
# Go Patterns (this repo)

> Repo-specific idioms — generic Go design patterns are covered by the
> globally-installed `ecc` plugin's `golang-patterns` skill.

## hotcereal graph/store

Domain types persisted via `github.com/paynejacob/hotcereal` (e.g.
`pkg/auth.Token`, `pkg/auth.User`, `pkg/sound.Sound`, `pkg/sound.Group`)
are plain Go structs with hotcereal struct tags controlling how fields are
indexed/stored:

- A field tagged as the key (e.g. `Id`) is the primary lookup key.
- Fields tagged `searchable` (e.g. `Sound.Name`) get a secondary index.
- Fields tagged `lazy` (e.g. `Sound.Audio []byte`) are stored/streamed
  separately from the main record via `store.ReadLazy`/`WriteLazy`,
  instead of being inlined into every read of the record.

Each such type has a generated `zz_<Type>_provider.go` sibling file (e.g.
`pkg/sound/zz_Sound_provider.go`) implementing CRUD (`Get`, `List`, `Save`,
`Delete`, ...) against a `hotcereal/pkg/store.Store`. **Never hand-edit
these files** — see `codegen.md`.

## Service pattern

Every domain package that exposes HTTP routes implements `pkg/service.Service`:

```go
type Service interface {
    RegisterRoutes(*mux.Router)
    Run(ctx context.Context)
}
```

`pkg/service.Manager` collects registered services, calls
`RegisterRoutes` immediately on registration, and runs each service's
`Run(ctx)` in its own goroutine (see `pkg/server/server.go`,
`pkg/service/service.go`). A new domain package should follow this same
shape rather than wiring routes/background loops ad hoc.

## Broadcast pattern

`pkg/websocket.Service.BroadcastMessage` pushes a JSON message to every
connected client's per-connection buffered channel. Domain services (see
`pkg/sound/service.go`) call this synchronously right after a DB write for
CRUD events (`update_sound`, `delete_sound`, ...), but "play" events are
enqueued onto `pkg/sound/queue.go`'s sequential `playQueue` instead, which
broadcasts only when it actually starts playback — so a `play` HTTP call
returning `202 Accepted` does not mean the broadcast happened yet.
```

- [ ] **Step 6: Create `.claude/rules/golang/codegen.md`**

```markdown
---
paths:
  - "pkg/**/*.go"
---
# Code Generation (hotcereal)

## The rule

**Never hand-edit a `zz_*_provider.go` file.** These are generated by
`go generate ./pkg/...` from the `hotcereal` struct tags on the
corresponding domain type (see `patterns.md`). If a provider's behavior
needs to change, change the struct/tags on the source type and regenerate
— don't patch the generated file directly, the next `go generate` run
will silently overwrite your edit.

## Workflow

1. Add or modify a `hotcereal`-tagged struct field on the domain type
   (e.g. `pkg/sound/sound.go`).
2. Run:
   ```bash
   go generate ./pkg/...
   ```
3. Verify the diff is what you expect — no unrelated files should change:
   ```bash
   git status --porcelain
   git diff
   ```
4. Run `scripts/validate/generate` to confirm CI's exact check passes
   (it fails the build on any uncommitted diff after `go generate`):
   ```bash
   scripts/validate/generate
   ```

## Reference

See skill: `speakerbob-codegen` for a worked example.
```

- [ ] **Step 7: Validate frontmatter and non-empty content**

Run:
```bash
for f in .claude/rules/common/coding-style.md .claude/rules/common/testing.md \
         .claude/rules/golang/coding-style.md .claude/rules/golang/testing.md \
         .claude/rules/golang/patterns.md .claude/rules/golang/codegen.md; do
  test -s "$f" && head -1 "$f" | grep -q '^---$' && echo "OK: $f" || echo "FAIL: $f"
done
```
Expected: six `OK:` lines, no `FAIL:` lines.

- [ ] **Step 8: Commit**

```bash
git add .claude/rules/common .claude/rules/golang
git commit -m "Add common and golang Claude rules"
```

---

### Task 2: Rules — frontend boundary

**Files:**
- Create: `.claude/rules/frontend/boundary.md`

**Interfaces:**
- Consumes: N/A.
- Produces: `.claude/rules/frontend/boundary.md` — referenced by Task 7's `CLAUDE.md`.

- [ ] **Step 1: Create `.claude/rules/frontend/boundary.md`**

```markdown
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
```

- [ ] **Step 2: Validate**

```bash
test -s .claude/rules/frontend/boundary.md && head -1 .claude/rules/frontend/boundary.md | grep -q '^---$' && echo OK
```
Expected: `OK`.

- [ ] **Step 3: Commit**

```bash
git add .claude/rules/frontend
git commit -m "Add frontend boundary Claude rule"
```

---

### Task 3: Mappings

**Files:**
- Create: `.claude/mappings/package-map.md`
- Create: `.claude/mappings/request-flow.md`
- Create: `.claude/mappings/build-release-map.md`

**Interfaces:**
- Consumes: N/A.
- Produces: the three mapping file paths — referenced by Task 7's `CLAUDE.md` and by `.claude/skills/*` and `.claude/rules/*` cross-references already written in Tasks 1–2.

- [ ] **Step 1: Create `.claude/mappings/package-map.md`**

```markdown
# Package Map

One entry per `pkg/*` package: responsibility, key types, dependencies.
Generated `zz_*_provider.go` files are omitted — see
`../rules/golang/codegen.md`.

## pkg/auth

OAuth-based authentication/authorization and session/token management.

- `Service` (`service.go`): holds `TokenProvider`, `UserProvider`, an
  in-memory `StateManager` (OAuth CSRF state), and the configured
  `[]Provider`s. `RegisterRoutes` registers `/user/preferences/`,
  `/login/`, `/logout/`, `/providers/`, `/callback/`, `/tokens/`,
  `/tokens/ws/`. `Handler(h)` wraps an `http.Handler` with auth
  enforcement (`VerifyRequest`). `Run(ctx)` runs a background goroutine
  that cleans up expired tokens every hour.
- `Provider` interface (`provider.go`): `Name()`, `VerifyCallback(r)`,
  `LoginRedirect(w, r, state)`. `pkg/auth/github` is the one concrete
  implementation in this repo (GitHub OAuth, org/email allowlists).
- `Token` (`token.go`): hotcereal type — `Session`, `Bearer`, or
  `Websocket` typed tokens, each with its own extraction rule from a
  request (cookie, `Authorization` header, or `?token=` query param
  respectively).
- `User` (`user.go`): hotcereal type — `Email`, `Principals`
  (`"<provider>://<id>"` strings), free-form `Preferences`.
- `StateManager` (`state.go`): in-memory map with a 5-minute
  self-expiring entry per OAuth flow.
- Depends on: `pkg/service` (error responses), `gorilla/mux`.

## pkg/auth/github

GitHub OAuth `auth.Provider` implementation (package name is `auth`,
directory is `github`). Exchanges an OAuth code for a token, then checks
the authenticated GitHub user's orgs and primary email against configured
allowlists (`OrganizationPermissionMap`, `EmailPermissionMap`); denies via
`auth.AccessDenied{}` otherwise.

**Known limitation**: the org-membership check fetches only the first 100
orgs (`per_page=100`) with no pagination — a user in >100 orgs can be
denied incorrectly.

## pkg/health

Trivial liveness endpoint. `GET /healthz` → `"ok"`. No dependencies beyond
`pkg/service`.

## pkg/server

The composition root. `NewServer` builds every hotcereal provider bound to
the given store, wires up the `mux.Router` (`/auth` and `/api`
subrouters — only `/api/*` requires auth, since auth's own routes must be
reachable unauthenticated), constructs and registers every domain
`Service` via `pkg/service.Manager`, and sets `pkg/static.Service{}` as
the router's `NotFoundHandler` (SPA fallback). `Run(ctx)` starts the
manager and blocks on `http.Server.ListenAndServe`, shutting down
gracefully on context cancellation (10s timeout).

## pkg/service

Shared plumbing, not a domain package. `Service` interface
(`RegisterRoutes`, `Run`) and `Manager` (registers + runs every service's
background goroutine). `errors.go` defines `SpeakerbobError`,
`NotAcceptableError`, and `WriteErrorResponse` — the repo's standard
error-to-HTTP-response mapping. See `../rules/golang/patterns.md` for the
service pattern new packages should follow.

## pkg/sound

The core domain: sound clips, groups, playback, and text-to-speech.

- `Sound` (`sound.go`): hotcereal type — searchable `Name`, `Duration`,
  `Hidden` (transient, not persisted to JSON responses), lazy `Audio
  []byte`. Created either from an upload (`NewSound`, normalizes via
  `ffmpeg`) or from text (`NewTTSSound`, via `flite`, always `Hidden`).
- `Group` (`group.go`): hotcereal type — a named, ordered `SoundIds`
  playlist.
- `audio.go`: shells out to `ffmpeg` (loudnorm, mp3, duration-capped) and
  `flite` (TTS) via `os/exec`. Requires both binaries on `PATH`
  (installed in the Docker image, see `build-release-map.md`).
- `queue.go`: a single global sequential `playQueue` — enqueue appends and
  signals a channel; a `ConsumeQueue` goroutine pops one sound at a time,
  broadcasts a `PlayMessage`, and waits out its `Duration` before playing
  the next. See `../rules/golang/patterns.md`'s "Broadcast pattern" —
  playback broadcasts are asynchronous relative to the HTTP call that
  enqueued them.
- `service.go`: registers `/sound/sounds/*`, `/sound/groups/*`,
  `/sound/search/`, `/sound/say/`. `Run(ctx)` starts `ConsumeQueue` and a
  4-hour ticker that deletes `Hidden` sounds older than 24h.
- `utils.go`: `DeleteSoundWithGroups` — deletes any group referencing a
  sound before deleting the sound itself (referential cleanup).
- Depends on: `pkg/service`, `pkg/websocket`, external `ffmpeg`/`flite`.

## pkg/static

Embeds the built frontend (`//go:embed assets`) and serves it as the
router's catch-all. Falls back to `index.html` for any unknown path (SPA
routing). `index.html` is served with `no-cache` (so shell updates
propagate immediately); versioned assets (`.js`/`.css`/`.woff2`/`.ico`)
get a 1-year cache header. See `../rules/frontend/boundary.md`.

## pkg/store/badgerdb

Adapter implementing `hotcereal/pkg/store.Store` over
`dgraph-io/badger/v3`. `Get`/`List`/`ReadLazy`/`WriteLazy`/`Save`/
`BulkSave`/`Delete`/`Close`. No custom transaction/retry logic beyond
badger's own `View`/`Update`. This is the only storage backend in the
repo today.

## pkg/version

Single exported `Version` string constant, overridden at build time via
`-ldflags -X .../pkg/version.Version=...` (see `build-release-map.md`).

## pkg/websocket

Broadcast-only websocket fan-out to connected browser clients.

- `Service` (`service.go`): `GET /ws/` upgrades a connection after
  `AuthService.VerifyWebsocket` (accepts `Bearer`, `Session`, or the
  short-lived `Websocket` token type obtained from
  `/auth/tokens/ws/`). `BroadcastMessage` pushes JSON to every
  connection's buffered send channel (size 256). Connect/disconnect each
  broadcast a live viewer-count message.
- `Conn` (`connection.go`): `readPump` exists only to enforce a 60s read
  deadline and detect disconnects — it discards all client-sent data,
  since clients never send meaningful messages on this connection.
  `writePump` drains the send channel and pings every 54s.
- Depends on: `pkg/auth` (for verification), `gorilla/websocket`.
```

- [ ] **Step 2: Create `.claude/mappings/request-flow.md`**

```markdown
# Request Flow

## Process startup

`main.go` → `cmd.Execute()` (cobra) → `cmd/server` command (`server.go`,
build-tag `!windows`) → loads YAML config (`config.go`; writes a default
config file on first run if none exists) → opens `badger.DB` at
`config.DataPath`, wraps it in `pkg/store/badgerdb.Store` → calls
`pkg/server.NewServer(store, config)` then `Run(ctx)` (ctx cancelled on
SIGINT).

`cmd/server/development.go` (build-tag `development`) swaps in a stub
`DevAuthProvider` that always authenticates as a fixed debug user, and
points `DataPath`/`Host`/`Port` at local dev defaults — this is how local
development runs without real OAuth configured.

## HTTP request: playing a sound

`PUT /api/sound/sounds/{soundId}/play/`:

1. `mux.Router` (built in `pkg/server.NewServer`) matches the `/api`
   subrouter.
2. `apiRouter.Use(authService.Handler)` middleware runs first —
   `pkg/auth.Handler.ServeHTTP` calls `VerifyRequest`; a missing/invalid
   `Bearer`/`Session` token short-circuits with `401`.
3. Routed to `pkg/sound.Service.playSound`.
4. `SoundProvider.Get(soundId)` reads the record through the generated
   provider, which reads from `pkg/store/badgerdb.Store` via hotcereal's
   key encoding.
5. `playQueue.EnqueueSounds(sound)` — responds `202 Accepted`
   immediately. Playback is asynchronous: see `pkg/sound/queue.go`'s
   `ConsumeQueue` goroutine, which broadcasts a `PlayMessage` only once it
   actually starts playing this sound (it may be queued behind another
   sound already playing).

CRUD-style requests (create/update/delete a sound or group) follow the
same auth-middleware step, then persist via the relevant provider and
broadcast the corresponding message (`update_sound`, `delete_sound`, ...)
**synchronously**, right after the write — unlike `play`, which goes
through the queue.

## Websocket path

1. Browser calls `GET /auth/tokens/ws/` (authenticated) to obtain a
   short-lived `Websocket`-type token.
2. Browser opens `GET /ws/?token=<that token>`. `pkg/websocket.Service`
   verifies it via `AuthService.VerifyWebsocket`, upgrades via
   `gorilla/websocket`, and registers a `Conn`.
3. From then on, any `BroadcastMessage` call (from `pkg/sound.Service`'s
   mutation handlers, or from `playQueue.ConsumeQueue` when a sound starts
   playing) pushes to every connection's buffered channel; each
   connection's own `writePump` goroutine drains its channel
   independently and writes JSON to that client.
4. Connect/disconnect events themselves broadcast a live viewer-count
   message to everyone.

This is a broadcast-only channel — the server never expects or processes
data sent *from* a client over the websocket (`Conn.readPump` discards
everything it reads; it exists solely to detect disconnects via read
timeouts).
```

- [ ] **Step 3: Create `.claude/mappings/build-release-map.md`**

```markdown
# Build & Release Map

How `scripts/`, `.github/workflows/`, `Dockerfile`, and `charts/` fit
together.

## Local validation (`scripts/validate/*`) — run by `validate.yaml` on every PR

| Script | What it checks |
|---|---|
| `golang-lint` | `go mod tidy && go fmt ./...`, fails the job on any resulting diff |
| `golang-test` | `go test ./...` |
| `generate` | `go generate ./pkg/...`, fails the job on any resulting diff — this is what enforces committing regenerated `zz_*.go` files, see `../rules/golang/codegen.md` |
| `web` | `yarn lint` in `web/speakerbob`, fails on any diff |
| `helm` | `helm lint charts/speakerbob` |
| `docker` | full `docker build --no-cache .` (no push) |

Each runs as its own parallel CI job, currently all pinned to Go
`1.16.6` via `actions/setup-go@v2` in `validate.yaml`.

## `scripts/version`

Derives `VERSION`/`IMAGE_TAG`/`CHART_VERSION`/`IS_PRERELEASE`/
`IMAGE_NAME` from git tags (exact tag match) or branch+short-hash
otherwise, exporting them as shell vars and (when `GITHUB_ACTIONS=true`)
into `$GITHUB_ENV` for later workflow steps.

## `scripts/build` — only runs in `release.yaml`

Builds the frontend (`yarn build`) and moves its output to repo-root
`assets/` — **this path is load-bearing**: `pkg/static/service.go`'s
`//go:embed assets` expects the built frontend there. Stamps
`charts/speakerbob/Chart.yaml` and `docs/{asyncapi,openapi}.yaml` version
fields via `yq`. Cross-compiles 3 binaries (linux/arm64, linux/amd64,
windows; `CGO_ENABLED=1` — needed for the `badger`/cgo dependencies —
with the version ldflag setting `pkg/version.Version`). Packages the helm
chart and copies API spec docs into `dist/`.

## `Dockerfile`

Multi-stage: `uibuild` (node/yarn, builds `web/speakerbob`) →
`gobuild` (`golang:1.16.6-alpine3.13`, copies the `uibuild` output to
`pkg/static/assets`, builds the Go binary with `CGO_ENABLED=1`) → final
`alpine` stage installing `ffmpeg`/`flite` (runtime dependencies of
`pkg/sound/audio.go`) and copying just the built binary. Used by both
`push-image.yaml` and `release.yaml`.

## `charts/speakerbob/`

Standard Helm chart layout (`Chart.yaml`, `values.yaml`,
`templates/{deployment,ingress,pvc,secret,service}.yaml`, etc.).
`Chart.yaml`'s `version`/`appVersion` are not hand-maintained — they're
overwritten by `scripts/build` at release time.

## GitHub Actions workflows

| Workflow | Trigger | Does |
|---|---|---|
| `validate.yaml` | pull request | runs all `scripts/validate/*` jobs (table above) |
| `push-image.yaml` | push to `master`/`release/*` | `scripts/version`, then builds+pushes **only** the Docker image to `ghcr.io` (tagged `$VERSION` and a branch-based `$IMAGE_TAG`) — no binaries, no chart packaging |
| `release.yaml` | push of a `v*` tag | `scripts/version`, `scripts/build` (binaries + packaged chart + API specs into `dist/`), builds+pushes the Docker image, uploads the packaged chart to an external chart repo, creates a GitHub release with `dist/*` as assets |

So: PR checks never build release artifacts; pushes to `master` only
refresh the running Docker image; a full versioned release (binaries,
Helm chart, GitHub release) only happens on a `v*` tag push.
```

- [ ] **Step 4: Validate**

```bash
for f in .claude/mappings/package-map.md .claude/mappings/request-flow.md .claude/mappings/build-release-map.md; do
  test -s "$f" && echo "OK: $f" || echo "FAIL: $f"
done
```
Expected: three `OK:` lines.

- [ ] **Step 5: Commit**

```bash
git add .claude/mappings
git commit -m "Add package, request-flow, and build-release mapping docs"
```

---

### Task 4: Learnings

**Files:**
- Create: `.claude/learnings/README.md`
- Create: `.claude/learnings/2026-09-11-codegen-provider-files.md`
- Create: `.claude/learnings/2026-09-11-frontend-out-of-scope.md`

**Interfaces:**
- Consumes: N/A.
- Produces: the three learning file paths — referenced by Task 7's `CLAUDE.md`.

- [ ] **Step 1: Create `.claude/learnings/README.md`**

```markdown
# Learnings

Atomic, git-committed knowledge entries about this repo — modeled on the
globally-installed `ecc` plugin's `continuous-learning-v2` "instinct"
format (one trigger, one action, confidence-scored, evidence-backed), but
as static files with no hooks/observer/CLI machinery — that runtime system
already exists globally via `ecc` if the user has it installed, and
reimplementing it here would be out of scope for a single project.

## Format

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

## Confidence scale

| Score | Meaning |
|---|---|
| 0.3 | Tentative — observed once, worth a note but not a hard rule |
| 0.5 | Moderate — apply when relevant, open to revision |
| 0.7 | Strong — apply by default |
| 0.9 | Near-certain — treat as a hard constraint |

## Adding entries

File name: `YYYY-MM-DD-<id>.md`. Add a new entry whenever the Go version
upgrade or code-improvement work (sub-projects B and C) turns up a real
gotcha — a dependency behaving differently on a newer Go version, a
codegen quirk, a test-infra decision that took more than one attempt to
get right. Don't add speculative entries for things that haven't actually
been observed.
```

- [ ] **Step 2: Create `.claude/learnings/2026-09-11-codegen-provider-files.md`**

```markdown
---
id: codegen-provider-files
trigger: "editing any file under pkg/*, or adding a new hotcereal-tagged type"
confidence: 0.9
domain: codegen
source: repo-exploration
scope: project
---
# zz_*_provider.go files are generated, never hand-edited

## Action

Never edit a `zz_*_provider.go` file directly (e.g.
`pkg/auth/zz_User_provider.go`, `pkg/sound/zz_Sound_provider.go`). Change
the source type's struct/tags instead and run `go generate ./pkg/...`.
See `../rules/golang/codegen.md` and skill `speakerbob-codegen`.

## Evidence

- `scripts/validate/generate` (run in CI's `validate.yaml`) runs
  `go generate ./pkg/...` and fails the build on any resulting diff —
  confirming these files are expected to always match generator output.
- Every `zz_*_provider.go` file in the repo imports
  `github.com/paynejacob/hotcereal/pkg/graph` and
  `github.com/paynejacob/hotcereal/pkg/store`, consistent with being
  generated by that library's codegen rather than hand-written.
```

- [ ] **Step 3: Create `.claude/learnings/2026-09-11-frontend-out-of-scope.md`**

```markdown
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
```

- [ ] **Step 4: Validate frontmatter on the two atomic entries**

```bash
for f in .claude/learnings/2026-09-11-codegen-provider-files.md .claude/learnings/2026-09-11-frontend-out-of-scope.md; do
  grep -q '^id:' "$f" && grep -q '^confidence:' "$f" && grep -q '^## Action$' "$f" && grep -q '^## Evidence$' "$f" && echo "OK: $f" || echo "FAIL: $f"
done
test -s .claude/learnings/README.md && echo "OK: README.md"
```
Expected: two `OK:` lines for the entries plus `OK: README.md`.

- [ ] **Step 5: Commit**

```bash
git add .claude/learnings
git commit -m "Add learnings scaffold and two seed entries"
```

---

### Task 5: Skill — speakerbob-codegen

**Files:**
- Create: `.claude/skills/speakerbob-codegen/SKILL.md`

**Interfaces:**
- Consumes: references `.claude/mappings/package-map.md`, `.claude/rules/golang/patterns.md`, `.claude/rules/golang/testing.md` (created in Tasks 1 and 3).
- Produces: skill name `speakerbob-codegen`, invokable via the `Skill` tool once `.claude/skills/` is discovered.

- [ ] **Step 1: Create `.claude/skills/speakerbob-codegen/SKILL.md`**

```markdown
---
name: speakerbob-codegen
description: Use when adding or modifying a hotcereal graph-backed domain type in this repo (e.g. pkg/sound.Sound, pkg/auth.Token) and regenerating its zz_*_provider.go file.
---

# speakerbob: hotcereal codegen workflow

This repo uses `github.com/paynejacob/hotcereal` to generate persistence
code (`zz_<Type>_provider.go`) for domain types from struct tags. See
`.claude/mappings/package-map.md` and `.claude/rules/golang/patterns.md`
for what these types look like in this repo.

## When to use this skill

- Adding a new field to an existing hotcereal-tagged type (e.g. adding a
  field to `pkg/sound.Sound`).
- Adding a brand-new hotcereal-tagged domain type.
- Any time a `zz_*_provider.go` file looks like it needs to change.

## Steps

1. **Never edit the `zz_*_provider.go` file directly.** Find the source
   type instead (e.g. `pkg/sound/sound.go` for `Sound`) and change its
   struct fields/tags there.

2. Regenerate:

   ```bash
   go generate ./pkg/...
   ```

3. Inspect the diff — only the relevant `zz_*_provider.go` file(s) should
   have changed:

   ```bash
   git status --porcelain
   git diff -- '*/zz_*_provider.go'
   ```

   If unrelated files changed, something is wrong — investigate before
   continuing (e.g. a stale generator, a tag typo affecting an unrelated
   type).

4. Confirm the generation step is clean the way CI checks it (this fails
   the build on any diff after running):

   ```bash
   scripts/validate/generate
   ```

5. Run tests for the package you changed (and any package that embeds or
   depends on it) before committing — see `.claude/rules/golang/testing.md`
   for how tests in this repo are structured.

## Common mistakes

- Editing the generated file to "quickly fix" a bug — the next
  `go generate` silently reverts it.
- Forgetting to run `go generate` after a struct change, so CI catches it
  later instead of you catching it locally.
```

- [ ] **Step 2: Validate frontmatter**

```bash
f=.claude/skills/speakerbob-codegen/SKILL.md
test -s "$f" && grep -q '^name: speakerbob-codegen$' "$f" && grep -q '^description:' "$f" && echo OK
```
Expected: `OK`.

- [ ] **Step 3: Invoke the skill once to confirm it loads**

Use the `Skill` tool with `skill: "speakerbob-codegen"` and confirm the
returned content matches what was written above.

- [ ] **Step 4: Commit**

```bash
git add .claude/skills/speakerbob-codegen
git commit -m "Add speakerbob-codegen skill"
```

---

### Task 6: Skill — speakerbob-validate

**Files:**
- Create: `.claude/skills/speakerbob-validate/SKILL.md`

**Interfaces:**
- Consumes: N/A.
- Produces: skill name `speakerbob-validate`, invokable via the `Skill` tool.

- [ ] **Step 1: Create `.claude/skills/speakerbob-validate/SKILL.md`**

```markdown
---
name: speakerbob-validate
description: Use before considering any Go change in this repo done — runs the same checks CI's validate.yaml runs locally (fmt/tidy diff, generate diff, go test).
---

# speakerbob: local validation (matches CI)

`validate.yaml` runs several `scripts/validate/*` checks on every PR. Run
the Go-relevant ones locally before calling a change done, so CI isn't the
first place a problem shows up.

## Steps

Run each of these from the repo root. Each should produce **no output and
exit 0** — any output or non-zero exit means something needs fixing
before committing.

1. **Format / tidy check** (mirrors `scripts/validate/golang-lint`):

   ```bash
   go mod tidy
   go fmt ./...
   git status --porcelain --untracked-files=no
   ```

   If the last command prints anything, `go mod tidy`/`go fmt` changed
   something — review and commit those changes.

2. **Codegen check** (mirrors `scripts/validate/generate`):

   ```bash
   go generate ./pkg/...
   git status --porcelain --untracked-files=no
   ```

   If this prints anything, a `zz_*_provider.go` file is out of date —
   see skill `speakerbob-codegen`.

3. **Test check** (mirrors `scripts/validate/golang-test`):

   ```bash
   go test ./...
   ```

## What this intentionally skips

- `scripts/validate/web`, `helm`, `docker` — these validate the frontend,
  Helm chart, and Docker build respectively. Only run them if the change
  actually touches `web/speakerbob/`, `charts/`, or the `Dockerfile`.
```

- [ ] **Step 2: Validate frontmatter**

```bash
f=.claude/skills/speakerbob-validate/SKILL.md
test -s "$f" && grep -q '^name: speakerbob-validate$' "$f" && grep -q '^description:' "$f" && echo OK
```
Expected: `OK`.

- [ ] **Step 3: Invoke the skill once to confirm it loads**

Use the `Skill` tool with `skill: "speakerbob-validate"` and confirm the
returned content matches what was written above.

- [ ] **Step 4: Commit**

```bash
git add .claude/skills/speakerbob-validate
git commit -m "Add speakerbob-validate skill"
```

---

### Task 7: Root CLAUDE.md and full-scaffold validation

**Files:**
- Create: `CLAUDE.md`

**Interfaces:**
- Consumes: every file path created in Tasks 1–4 (rules, mappings, learnings) via `@`-import; references skill names `speakerbob-codegen` and `speakerbob-validate` from Tasks 5–6 in prose (not `@`-imported — skills are auto-discovered, not imported).
- Produces: root `CLAUDE.md`, completing the scaffold.

- [ ] **Step 1: Create `CLAUDE.md`**

```markdown
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
go generate ./pkg/...   # regenerate hotcereal provider code — see rules/golang/codegen.md
```

Before considering a Go change done, run the checks CI runs — see skill
`speakerbob-validate`.

## Rules

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
```

- [ ] **Step 2: Verify every `@`-import path resolves**

```bash
grep '^@' CLAUDE.md | sed 's/^@//' | while read -r p; do
  test -f "$p" && echo "OK: $p" || echo "MISSING: $p"
done
```
Expected: 13 `OK:` lines (7 rules + 3 mappings + 3 learnings), no
`MISSING:` lines.

- [ ] **Step 3: Full scaffold structural validation**

```bash
# Every SKILL.md has name + description and non-empty body
for f in .claude/skills/*/SKILL.md; do
  test -s "$f" && grep -q '^name:' "$f" && grep -q '^description:' "$f" && echo "OK: $f" || echo "FAIL: $f"
done

# No existing tracked file was modified by this plan
git status --porcelain --untracked-files=no
```
Expected: `OK:` for both `SKILL.md` files, no `FAIL:` lines, and the
`git status` line produces **no output** (this task only adds new files;
if it prints anything, an existing tracked file was unexpectedly
modified — stop and investigate before committing).

- [ ] **Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "Add root CLAUDE.md importing rules, mappings, and learnings"
```

---

## Self-Review Notes

- **Spec coverage**: every file listed in the spec's directory layout has
  a corresponding task/step (Tasks 1–7 cover all 16 files: 6 rules
  common/golang + 1 frontend rule + 3 mappings + 3 learnings + 2 skills +
  1 root `CLAUDE.md`).
- **Placeholder scan**: no "TBD"/"TODO placeholder" content — every step
  has real, final content.
- **Type consistency**: skill names (`speakerbob-codegen`,
  `speakerbob-validate`) match exactly between their `SKILL.md`
  frontmatter, their directory names, and every cross-reference in rules/
  learnings/`CLAUDE.md`. File paths in `@`-imports match exactly what
  Tasks 1–4 create.
