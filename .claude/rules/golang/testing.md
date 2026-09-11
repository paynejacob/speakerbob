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
