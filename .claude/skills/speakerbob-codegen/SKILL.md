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
