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
