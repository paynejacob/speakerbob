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
