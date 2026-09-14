# Go Version Upgrade — Design

## Context

Sub-project B of the repo modernization effort (sub-project A, the `.claude/`
tooling scaffold, is complete — see PR
https://github.com/stevemcquaid/speakerbob/pull/1). This spec covers
upgrading the Go toolchain and dependency set. Sub-project C (incremental
code/test-quality improvements) remains separate and later.

Today's date is 2026-09-11. Go's actual release/support state (verified via
`go.dev/doc/devel/release`, not assumed from training-data recall) is:

| Version | Released | Status |
|---|---|---|
| Go 1.27 | Aug 2026 | Current stable |
| Go 1.26 | Feb 2026 | Supported |
| Go 1.25 | Aug 2025 | Supported (oldest of the 3) |
| Go 1.24 / 1.23 / 1.22 | 2024–2025 | EOL |

Go supports only the latest 3 major releases at a time. The user chose
**Go 1.26** (latest patch verified via `go.dev/dl/?mode=json&include=all`:
**1.26.8**) after an earlier round where 1.22/1.23 were considered and
found to already be EOL.

## Current state

- `go.mod`: `go 1.16`, module `github.com/paynejacob/speakerbob`.
- CI (`.github/workflows/validate.yaml`, `release.yaml`) pins Go `1.16.6`
  via `actions/setup-go@v2` on 3 jobs (`go-fmt`/`golang-lint`,
  `go-test`/`golang-test`, `generate`) in `validate.yaml`, plus one job in
  `release.yaml`. Both workflows also pin several other GitHub Actions at
  very old majors (`actions/checkout@v2`, `actions/setup-node@v2`,
  `docker/login-action@v1`, `softprops/action-gh-release@v1`) —
  spot-checked two: `actions/checkout` is now at v7.0.1, `actions/setup-go`
  at v7.0.0 (both verified live via the GitHub API), so these are ~5 majors
  stale.
- `Dockerfile`: builder stage `golang:1.16.6-alpine3.13`, final runtime
  stage `alpine:3.13`. Verified via Docker Hub that `golang:1.26.8-alpine`
  currently resolves to `alpine3.24`.
- Two files use legacy `// +build` tags with no `//go:build` line:
  `cmd/server/server.go` (`!windows`), `cmd/server/development.go`
  (`development`). No other `.go` file in the repo uses build tags.
- Direct dependencies and their latest same-major versions (queried live
  via `go list -m -versions <module>`, not assumed):

  | Module | Current | Latest same-major |
  |---|---|---|
  | `github.com/dgraph-io/badger/v3` | v3.2011.1 | v3.2103.5 |
  | `github.com/gavv/httpexpect/v2` | v2.3.1 | v2.17.0 |
  | `github.com/google/uuid` | v1.2.0 | v1.6.0 |
  | `github.com/gorilla/handlers` | v1.5.1 | v1.5.2 |
  | `github.com/gorilla/mux` | v1.8.0 | v1.8.1 |
  | `github.com/gorilla/websocket` | v1.4.2 | v1.5.3 |
  | `github.com/sirupsen/logrus` | v1.7.0 | v1.10.2 |
  | `github.com/spf13/cobra` | v1.2.1 | v1.10.2 |
  | `github.com/stretchr/testify` | v1.7.0 | v1.12.1 |
  | `github.com/vmihailenco/msgpack/v5` | v5.3.4 | v5.4.1 |
  | `gopkg.in/yaml.v2` | v2.4.0 | v2.4.0 (already latest) |
  | `golang.org/x/sys` (indirect) | pinned old pseudo-version | resolved automatically by `go mod tidy` |
  | `github.com/paynejacob/hotcereal` | pinned to a specific commit past `alpha8` (bumped in the immediately preceding commit, `30b145b`) | no newer tag exists (`alpha4`..`alpha8` are all the tags); leave as-is unless it proves incompatible |

  `dgraph-io/badger` also has a `v4` line (up to v4.9.6) — the user
  explicitly chose to **stay on v3** (bump to v3.2103.5 only). A v3→v4
  migration is a separate, larger, riskier effort (new import path, likely
  API changes in `pkg/store/badgerdb/store.go`, and that package currently
  has zero test coverage) — out of scope here, not silently done.

- Checked for the Go 1.22 for-loop-variable-capture semantics change (the
  one behavior-affecting language change in this version range): every
  `go func()` launched inside a loop in this codebase was inspected.
  `pkg/service/service.go:31` already uses the pre-1.22 manual-capture
  workaround (`svc := m.services[i]` before the goroutine). No other
  loop+goroutine pattern exists in non-test code. **No risk found.**

## Scope

1. `go.mod`: bump `go 1.16` → `go 1.26`. Bump the 10 direct dependencies
   listed above to their latest same-major version (skip `yaml.v2`,
   already latest; skip `hotcereal`, no newer tag; leave `x/sys` to
   automatic resolution). Regenerate `go.sum` via `go mod tidy`.
2. Fix the 2 legacy build-tag files: run `gofmt` under the new Go 1.26
   toolchain, which will add the modern `//go:build` line above each
   existing `// +build` comment (comment-only change, zero behavior
   change) — this is what `scripts/validate/golang-lint` will otherwise
   fail on for any future touch to either file, and it's the direct
   consequence of the toolchain bump, not scope creep.
3. CI: bump `actions/setup-go` to `1.26.8` (all 4 occurrences: 3 in
   `validate.yaml`, 1 in `release.yaml`). Bump `actions/checkout`,
   `actions/setup-node`, `docker/login-action`, and
   `softprops/action-gh-release` to their current major versions — look
   up the exact current tag for each via `gh api
   repos/<owner>/<repo>/releases/latest` at implementation time rather
   than hardcoding a number now that could already be stale by the time
   this is implemented.
4. `Dockerfile`: bump the builder stage to `golang:1.26.8-alpine3.24` and
   the final runtime stage to `alpine:3.24` (kept in sync with the builder
   stage — `alpine:3.13`'s package repositories are old enough that
   `apk add` may already fail against them, independent of anything else
   in this change).

## Verification approach

No new business logic is introduced by this sub-project, so there is no
new code to unit-test. "Sufficient test coverage while maintaining
existing functionality" here means: the existing test suite, build, and
vet all continue to pass against the *actual* target toolchain, not just
whatever's ambient locally.

- Install the exact pinned version locally rather than relying on the
  ambient toolchain (currently 1.25.7, close but not identical to the
  1.26.8 target): `go install golang.org/dl/go1.26.8@latest && go1.26.8
  download`, then run `go1.26.8 build ./...`, `go1.26.8 vet ./...`,
  `go1.26.8 test ./...` before committing each task.
- `go1.26.8 test ./...` is expected to still show the 2 pre-existing
  `pkg/sound` failures (`TestCreateSound`, `TestSay`) on this machine,
  since it lacks `ffmpeg`/`flite` — already documented in
  `.claude/learnings/`. Confirm via `git diff --stat -- '*.go'` on each
  task that no application logic changed, so any *new* test failure is
  immediately attributable to that task's own change.
- `go vet ./...` clean is the primary signal for dependency-bump
  breakage (catches API-incompatible call sites without needing Docker
  or CI to find out).
- Docker build validation: attempt `docker build --no-cache .` locally if
  Docker is available in the execution environment; if not, this must be
  called out explicitly as unverified rather than assumed to work — do
  not claim Dockerfile correctness without either running the build or
  saying plainly that it wasn't run.

## Task breakdown (for the implementation plan)

Staged, each independently reviewable and testable, matching how
sub-project A was executed (subagent-driven-development, one task per
unit, task-scoped review after each, final whole-branch review at the
end):

1. Go toolchain bump: `go.mod`'s `go` directive + the 2 build-tag file
   fixes. Verify with `go1.26.8 build/vet/test`.
2. Dependency bumps: all 10 direct dependencies + `go mod tidy` to
   regenerate `go.sum`. Verify with `go1.26.8 build/vet/test` again,
   specifically checking for any compile break from an API change.
3. CI workflow version bumps (`actions/setup-go` to `1.26.8`, plus the 4
   other stale Actions to their current majors).
4. `Dockerfile` base image bumps (builder + final stage), verified via a
   local `docker build` if available.

## Non-goals

- No badger v3→v4 migration (explicit user decision — separate future
  effort with test coverage added first).
- No changes to `hotcereal`'s pinned version beyond what's already
  committed (`30b145b`), unless it proves incompatible with Go 1.26 (not
  observed yet — if it is, that becomes a new finding to surface, not a
  silent extra change).
- No new unit tests for existing untested packages (sub-project C's
  concern, and even there scoped to touched code only) — this
  sub-project touches no application logic.
- No changes to `web/speakerbob/` (frontend) or to `scripts/build`'s
  already-documented embed-path bug
  (`.claude/learnings/2026-09-11-scripts-build-embed-path-mismatch.md`) —
  that's a separate, already-flagged issue, not part of this version
  bump.

## Post-implementation notes

Two things were adjudicated during actual implementation, not anticipated in the original design above:

1. **`cmd/server/server.go:58` needed a real code fix, not just the build-tag comment.** `go vet` under Go 1.26 flagged a pre-existing unkeyed struct literal (`store.TypeKey{"versionVersion", 7, 7}`) — this check isn't new to 1.26, it was simply never run before (CI never runs `go vet`). Fixed to `store.TypeKey{Body: "versionVersion", PackageLength: 7, TypeLength: 7}` (same three values, keyed — zero behavior change, verified against `hotcereal/pkg/store.TypeKey`'s actual field names).
2. **`go.mod`'s `go` directive ended up as `1.26.0`, not bare `1.26`.** `hotcereal`'s codegen tool (invoked via `//go:generate`, which `go mod tidy` can't see) needed `golang.org/x/tools` bumped past its old `v0.1.5` floor to keep working under a modern Go toolchain; the version that fixes it (`v0.50.0`) itself requires `go >= 1.26.0`, so `go mod tidy` enforces that as our module's floor too, on every run. `1.26.0` and `1.26` express the identical minimum language version — this is Go's own tooling requiring precision, not a scope deviation.
