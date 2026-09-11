---
paths:
  - "pkg/**/*.go"
---
# Go Patterns (this repo)

> Repo-specific idioms — generic Go design patterns are covered by the
> globally-installed `ecc` plugin's `ecc:golang-patterns` skill.

## hotcereal graph/store

Domain types persisted via `github.com/paynejacob/hotcereal` (e.g.
`pkg/auth.Token`, `pkg/auth.User`, `pkg/sound.Sound`, `pkg/sound.Group`)
are plain Go structs with hotcereal struct tags controlling how fields are
indexed/stored:

- A field tagged as the key (e.g. `Id`) is the primary lookup key.
- Fields tagged `searchable` (e.g. `Sound.Name`) get a secondary index.
- Fields tagged `lookup` (e.g. `Token.Token`, `User.Email`,
  `User.Principals`) generate a dedicated accessor on the provider (e.g.
  `GetByToken`, `GetByEmail`, `GetByPrincipals`) for looking the record up
  by that field's value instead of its primary key.
- Fields tagged `lazy` (e.g. `Sound.Audio []byte`) are stored/streamed
  separately from the main record via `store.ReadLazy`/`WriteLazy`,
  instead of being inlined into every read of the record.

Each such type has a generated `zz_<Type>_provider.go` sibling file (e.g.
`pkg/sound/zz_Sound_provider.go`) implementing CRUD (`Get`, `List`, `Save`,
`Delete`, ...) against a `hotcereal/pkg/store.Store`. **Never hand-edit
these files** — see `codegen.md`.

## Service pattern

Every domain package that exposes HTTP routes implements `pkg/service.Service`:

```go
type Service interface {
    RegisterRoutes(*mux.Router)
    Run(ctx context.Context)
}
```

`pkg/service.Manager` collects registered services, calls
`RegisterRoutes` immediately on registration, and runs each service's
`Run(ctx)` in its own goroutine (see `pkg/server/server.go`,
`pkg/service/service.go`). A new domain package should follow this same
shape rather than wiring routes/background loops ad hoc.

## Broadcast pattern

`pkg/websocket.Service.BroadcastMessage` pushes a JSON message to every
connected client's per-connection buffered channel. Domain services (see
`pkg/sound/service.go`) call this synchronously right after a DB write for
CRUD events (`update_sound`, `delete_sound`, ...), but "play" events are
enqueued onto `pkg/sound/queue.go`'s sequential `playQueue` instead, which
broadcasts only when it actually starts playback — so a `play` HTTP call
returning `202 Accepted` does not mean the broadcast happened yet.
