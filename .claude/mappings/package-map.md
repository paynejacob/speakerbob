# Package Map

One entry per `pkg/*` package: responsibility, key types, dependencies.
Generated `zz_*_provider.go` files are omitted — see
`../rules/golang/codegen.md`.

## pkg/auth

OAuth-based authentication/authorization and session/token management.

- `Service` (`service.go`): holds `TokenProvider`, `UserProvider`, an
  in-memory `StateManager` (OAuth CSRF state), and the configured
  `[]Provider`s. `RegisterRoutes` registers `/user/preferences/`,
  `/login/`, `/logout/`, `/providers/`, `/callback/`, `/tokens/`,
  `/tokens/ws/`. `Handler(h)` wraps an `http.Handler` with auth
  enforcement (`VerifyRequest`). `Run(ctx)` runs a background goroutine
  that cleans up expired tokens every hour.
- `Provider` interface (`provider.go`): `Name()`, `VerifyCallback(r)`,
  `LoginRedirect(w, r, state)`. `pkg/auth/github` is the one concrete
  implementation in this repo (GitHub OAuth, org/email allowlists).
- `Token` (`token.go`): hotcereal type — `Session`, `Bearer`, or
  `Websocket` typed tokens, each with its own extraction rule from a
  request (cookie, `Authorization` header, or `?token=` query param
  respectively).
- `User` (`user.go`): hotcereal type — `Email`, `Principals`
  (`"<provider>://<id>"` strings), free-form `Preferences`.
- `StateManager` (`state.go`): in-memory map with a 5-minute
  self-expiring entry per OAuth flow.
- Depends on: `pkg/service` (error responses), `gorilla/mux`.

## pkg/auth/github

GitHub OAuth `auth.Provider` implementation (package name is `auth`,
directory is `github`). Exchanges an OAuth code for a token, then checks
the authenticated GitHub user's orgs and primary email against configured
allowlists (`OrganizationPermissionMap`, `EmailPermissionMap`); denies via
`auth.AccessDenied{}` otherwise.

**Known limitation**: the org-membership check fetches only the first 100
orgs (`per_page=100`) with no pagination — a user in >100 orgs can be
denied incorrectly.

## pkg/health

Trivial liveness endpoint. `GET /healthz` → `"ok"`. No dependencies beyond
`pkg/service`.

## pkg/server

The composition root. `NewServer` builds every hotcereal provider bound to
the given store, wires up the `mux.Router` (`/auth` and `/api`
subrouters — only `/api/*` requires auth, since auth's own routes must be
reachable unauthenticated), constructs and registers every domain
`Service` via `pkg/service.Manager`, and sets `pkg/static.Service{}` as
the router's `NotFoundHandler` (SPA fallback). `Run(ctx)` starts the
manager and blocks on `http.Server.ListenAndServe`, shutting down
gracefully on context cancellation (10s timeout).

## pkg/service

Shared plumbing, not a domain package. `Service` interface
(`RegisterRoutes`, `Run`) and `Manager` (registers + runs every service's
background goroutine). `errors.go` defines `SpeakerbobError`,
`NotAcceptableError`, and `WriteErrorResponse` — the repo's standard
error-to-HTTP-response mapping. See `../rules/golang/patterns.md` for the
service pattern new packages should follow.

## pkg/sound

The core domain: sound clips, groups, playback, and text-to-speech.

- `Sound` (`sound.go`): hotcereal type — searchable `Name`, `Duration`,
  `Hidden` (transient, not persisted to JSON responses), lazy `Audio
  []byte`. Created either from an upload (`NewSound`, normalizes via
  `ffmpeg`) or from text (`NewTTSSound`, via `flite`, always `Hidden`).
- `Group` (`group.go`): hotcereal type — a named, ordered `SoundIds`
  playlist.
- `audio.go`: shells out to `ffmpeg` (loudnorm, mp3, duration-capped) and
  `flite` (TTS) via `os/exec`. Requires both binaries on `PATH`
  (installed in the Docker image, see `build-release-map.md`).
- `queue.go`: a single global sequential `playQueue` — enqueue appends and
  signals a channel; a `ConsumeQueue` goroutine pops one sound at a time,
  broadcasts a `PlayMessage`, and waits out its `Duration` before playing
  the next. See `../rules/golang/patterns.md`'s "Broadcast pattern" —
  playback broadcasts are asynchronous relative to the HTTP call that
  enqueued them.
- `service.go`: registers `/sound/sounds/*`, `/sound/groups/*`,
  `/sound/search/`, `/sound/say/`. `Run(ctx)` starts `ConsumeQueue` and a
  4-hour ticker that deletes `Hidden` sounds older than 24h.
- `utils.go`: `DeleteSoundWithGroups` — deletes any group referencing a
  sound before deleting the sound itself (referential cleanup).
- Depends on: `pkg/service`, `pkg/websocket`, external `ffmpeg`/`flite`.

## pkg/static

Embeds the built frontend (`//go:embed assets`) and serves it as the
router's catch-all. Falls back to `index.html` for any unknown path (SPA
routing). `index.html` is served with `no-cache` (so shell updates
propagate immediately); versioned assets (`.js`/`.css`/`.woff2`/`.ico`)
get a 1-year cache header. See `../rules/frontend/boundary.md`.

## pkg/store/badgerdb

Adapter implementing `hotcereal/pkg/store.Store` over
`dgraph-io/badger/v3`. `Get`/`List`/`ReadLazy`/`WriteLazy`/`Save`/
`BulkSave`/`Delete`/`Close`. No custom transaction/retry logic beyond
badger's own `View`/`Update`. This is the only storage backend in the
repo today.

## pkg/version

Single exported `Version` string constant, overridden at build time via
`-ldflags -X .../pkg/version.Version=...` (see `build-release-map.md`).

## pkg/websocket

Broadcast-only websocket fan-out to connected browser clients.

- `Service` (`service.go`): `GET /ws/` upgrades a connection after
  `AuthService.VerifyWebsocket` (accepts `Bearer`, `Session`, or the
  short-lived `Websocket` token type obtained from
  `/auth/tokens/ws/`). `BroadcastMessage` pushes JSON to every
  connection's buffered send channel (size 256). Connect/disconnect each
  broadcast a live viewer-count message.
- `Conn` (`connection.go`): `readPump` exists only to enforce a 60s read
  deadline and detect disconnects — it discards all client-sent data,
  since clients never send meaningful messages on this connection.
  `writePump` drains the send channel and pings every 54s.
- Depends on: `pkg/auth` (for verification), `gorilla/websocket`.
