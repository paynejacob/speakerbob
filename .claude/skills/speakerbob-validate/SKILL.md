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
