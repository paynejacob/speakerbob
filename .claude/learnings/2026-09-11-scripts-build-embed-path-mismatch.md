---
id: scripts-build-embed-path-mismatch
trigger: "touching scripts/build, pkg/static/service.go, or release.yaml's cross-compiled binaries"
confidence: 0.7
domain: build
source: code-improvement
scope: project
---
# scripts/build likely embeds a stale frontend placeholder, not the real build

## Action

Before changing `scripts/build`, `pkg/static/service.go`, or the release
binary pipeline, verify whether the cross-compiled binaries in
`release.yaml` actually ship a working frontend. `pkg/static/service.go`'s
`//go:embed assets` is package-relative (`pkg/static/assets/`), but
`scripts/build` moves the built frontend to repo-root `assets/`
(`mv dist ../../assets`, run from `web/speakerbob/`) — a different
location. `pkg/static/assets/` only contains a committed 75-byte
placeholder `index.html`. The `Dockerfile`'s `uibuild` stage sidesteps
this by copying directly to `pkg/static/assets` (`COPY --from=uibuild
/ui/dist pkg/static/assets`), so the Docker image is likely fine — but
the 3 binaries `scripts/build` cross-compiles for `release.yaml` may ship
without a working frontend. Not yet root-caused by running the pipeline
end-to-end — treat as a likely bug to confirm, not a confirmed one.

## Evidence

- `pkg/static/service.go:16`: `//go:embed assets` (package-relative).
- `pkg/static/assets/index.html` is 75 bytes, clearly a placeholder, not a
  built Vue app.
- `scripts/build:5-8`: `pushd web/speakerbob`, `yarn build`, `mv dist
  ../../assets`, `popd` — resolves to repo-root `assets/`, not
  `pkg/static/assets/`.
- `Dockerfile`'s `gobuild` stage: `COPY --from=uibuild /ui/dist
  pkg/static/assets` — correct path, but this is a separate build path
  from `scripts/build` (Docker builds the frontend itself in its
  `uibuild` stage, it doesn't invoke `scripts/build`).
