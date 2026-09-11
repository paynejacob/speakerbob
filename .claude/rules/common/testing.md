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
