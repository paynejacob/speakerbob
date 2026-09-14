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
