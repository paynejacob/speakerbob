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
- For storage-layer tests, use `hotcereal`'s in-memory store implementation
  (`github.com/paynejacob/hotcereal/pkg/stores/memory`, `memory.New()`)
  rather than mocking `hotcereal/pkg/store.Store` or wiring up a real
  `badgerdb.Store` — the existing `pkg/sound/service_test.go` is the
  reference example for this pattern.

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

## Runtime prerequisites for tests

`go test ./...` requires `ffmpeg` and `flite` on `PATH` — `pkg/sound`'s
tests shell out to both (see `pkg/sound/audio.go`). CI installs them
explicitly (`sudo apt install -y flite ffmpeg`) before running tests; do
the same locally, or those tests will fail with an exec error, not a Go
compile error.

## Reference

See the globally-installed `ecc` plugin skill `ecc:golang-testing` for
generic Go testing idioms not specific to this repo.
