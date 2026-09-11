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
