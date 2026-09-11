# Request Flow

## Process startup

`main.go` → `cmd.Execute()` (cobra) → `cmd/server` command (`server.go`,
build-tag `!windows`) → loads YAML config (`config.go`; writes a default
config file on first run if none exists) → opens `badger.DB` at
`config.DataPath`, wraps it in `pkg/store/badgerdb.Store` → calls
`pkg/server.NewServer(store, config)` then `Run(ctx)` (ctx cancelled on
SIGINT).

`cmd/server/development.go` (build-tag `development`) swaps in a stub
`DevAuthProvider` that always authenticates as a fixed debug user, and
points `DataPath`/`Host`/`Port` at local dev defaults — this is how local
development runs without real OAuth configured.

## HTTP request: playing a sound

`PUT /api/sound/sounds/{soundId}/play/`:

1. `mux.Router` (built in `pkg/server.NewServer`) matches the `/api`
   subrouter.
2. `apiRouter.Use(authService.Handler)` middleware runs first —
   `pkg/auth.Handler.ServeHTTP` calls `VerifyRequest`; a missing/invalid
   `Bearer`/`Session` token short-circuits with `401`.
3. Routed to `pkg/sound.Service.playSound`.
4. `SoundProvider.Get(soundId)` reads the record through the generated
   provider, which reads from `pkg/store/badgerdb.Store` via hotcereal's
   key encoding.
5. `playQueue.EnqueueSounds(sound)` — responds `202 Accepted`
   immediately. Playback is asynchronous: see `pkg/sound/queue.go`'s
   `ConsumeQueue` goroutine, which broadcasts a `PlayMessage` only once it
   actually starts playing this sound (it may be queued behind another
   sound already playing).

CRUD-style requests (create/update/delete a sound or group) follow the
same auth-middleware step, then persist via the relevant provider and
broadcast the corresponding message (`update_sound`, `delete_sound`, ...)
**synchronously**, right after the write — unlike `play`, which goes
through the queue.

## Websocket path

1. Browser calls `GET /auth/tokens/ws/` (authenticated) to obtain a
   short-lived `Websocket`-type token.
2. Browser opens `GET /ws/?token=<that token>`. `pkg/websocket.Service`
   verifies it via `AuthService.VerifyWebsocket`, upgrades via
   `gorilla/websocket`, and registers a `Conn`.
3. From then on, any `BroadcastMessage` call (from `pkg/sound.Service`'s
   mutation handlers, or from `playQueue.ConsumeQueue` when a sound starts
   playing) pushes to every connection's buffered channel; each
   connection's own `writePump` goroutine drains its channel
   independently and writes JSON to that client.
4. Connect/disconnect events themselves broadcast a live viewer-count
   message to everyone.

This is a broadcast-only channel — the server never expects or processes
data sent *from* a client over the websocket (`Conn.readPump` discards
everything it reads; it exists solely to detect disconnects via read
timeouts).
