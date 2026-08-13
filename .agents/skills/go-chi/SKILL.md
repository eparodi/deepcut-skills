---
name: go-chi
description: Go/chi backend development standards — layering, error handling, concurrency, state management, testing, database patterns, and common AI traps. Load when writing Go code in the backend/ directory.
---

# Go/chi Backend Standards

You are writing Go code using net/http stdlib + chi router v5. Follow
these conventions exactly. When your instinct (or a plausible-looking
API) contradicts this document, trust this document. These rules distill
the Uber Go Style Guide, Go Code Review Comments, and hard-won project
retros.

## Project Layout

**Rule: discover the layout before creating files.** Never assume a
layout — different Go projects use different structures, and creating
files in the wrong place is one of the most common AI failures.

```bash
# Discover where handlers/services/repos actually live:
find backend -name "*.go" -not -name "*_test.go" | head -30
```

Two common layouts you will encounter:

```
# A. Flat layered                    # B. Hexagonal modules
backend/                             backend/
├── cmd/server/main.go               ├── cmd/server/main.go
├── internal/                        ├── internal/
│   ├── handler/    # HTTP           │   ├── modules/<feature>/
│   ├── service/    # business       │   │   ├── domain/       # entities + ports
│   ├── store/      # data access    │   │   ├── application/  # services
│   ├── model/      # domain types   │   │   └── adapter/
│   ├── errs/       # error types    │   │       ├── http/     # handlers
│   └── config/                      │   │       └── postgres/ # repos
└── db/migrations/                   │   └── shared/{errs,render}/
                                     └── db/migrations/
```

New files go where equivalent files already live. If the repo uses
hexagonal modules, a new feature gets a new module directory — do NOT
introduce a parallel `handlers/` tree (or vice versa).

### Layer Rules (apply to any layout)

| Layer | Responsibility | Must NOT |
|-------|----------------|----------|
| Handler (HTTP adapter) | Parse/validate requests, render responses | Open DB connections, contain business rules |
| Service (application) | Business rules, orchestration | Import `net/http`, parse request bodies |
| Repository (store/adapter) | Database and external-API access | Import handlers or services |
| Domain (model) | Entities, typed constants, ports (interfaces) | Import any other layer |

**Rule:** validation of HTTP input (required fields, enum membership,
lengths) lives in the handler. Business rules (authorization, state
transitions) live in the service.

## Logging

**Rule:** Services MUST use an injected `*slog.Logger`, never the global
`slog` package functions — including in files that live NEXT TO code
that already does it right. Global calls bypass the configured handler
and make test output uncontrollable.

```go
// ✅ Right: injected logger with nil-safe helpers (tests pass nil)
type MyService struct {
    logger *slog.Logger
}

func (s *MyService) infoLog(msg string, args ...any) {
    if s.logger != nil {
        s.logger.Info(msg, args...)
    }
}

// ❌ Wrong: global slog inside a type that HAS a logger field
func (s *MyService) doWork() {
    slog.Info("work done") // bypasses configured handler/level
}
```

**Before finishing any service change:** `grep -n "slog\." <file>` and
replace stray global calls with the injected logger.

`main.go` MUST support `LOG_LEVEL` (`debug`, `info` default, `warn`,
`error`):

```go
func newLogger() *slog.Logger {
    level := slog.LevelInfo
    switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
    case "debug":
        level = slog.LevelDebug
    case "warn":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    }
    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
```

## State & Mutable Globals

**Package-level mutable state is banned** (Uber Go: Avoid Mutable
Globals). `var cache sync.Map` or `var seen = map[string]bool{}` at
package level is shared across every instance of your service — tests
can't isolate it, a second instance corrupts the first, and any future
writer can poison the type.

```go
// ❌ Wrong — package-level registry
var activeJobs sync.Map // map[string]context.CancelFunc

// ✅ Right — a field on the owning service
type JobService struct {
    // activeJobs tracks running jobs (jobID → context.CancelFunc).
    activeJobs sync.Map
}
```

**Type assertions from `any` use comma-ok — always.** A bare
`x.(string)` panics on mismatch; the map above is exactly where a future
refactor changes the stored type.

```go
// ❌ panics if the entry isn't a string
path := stored.(string)

// ✅ comma-ok, degrade gracefully
path, ok := stored.(string)
if !ok {
    s.warnLog("unexpected entry type", "key", key)
    return ""
}
```

`sync.Map` tip: use `LoadAndDelete` for pop semantics instead of
`Load` + `Delete` (atomic, less code).

## Enums & Magic Strings

Status/state strings written to the database or compared in code MUST be
typed constants in the domain package. Raw literals scattered across
call sites ("live", "processing", "failed") drift, typo silently, and
cannot be found by grep-for-type.

```go
// domain package
type JobStatus string

const (
    JobStatusPending JobStatus = "pending"
    JobStatusRunning JobStatus = "running"
    JobStatusFailed  JobStatus = "failed"
)

// call site
repo.UpdateStatus(ctx, id, string(domain.JobStatusFailed))
```

Inline literals inside SQL text (`WHERE status = 'live'`) are fine —
they are part of the query, not Go-side magic values. If two modules
must stay dependency-free, each defines its own constants with matching
values and a comment linking them.

Related magic-value rules:
- `"GET"` / `"POST"` → `http.MethodGet` / `http.MethodPost`
- Repeated buffer sizes / close codes → named constants
- `fmt.Errorf("static message")` with no verbs → `errors.New(...)`

## chi Router Patterns

### DO — Standard route registration

```go
func RegisterRoutes(r chi.Router, h *Handlers) {
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    r.Route("/api", func(r chi.Router) {
        r.Route("/users", func(r chi.Router) {
            r.Get("/", h.ListUsers)
            r.Post("/", h.CreateUser)
            r.Route("/{userID}", func(r chi.Router) {
                r.Use(h.UserCtx)        // loads user into context
                r.Get("/", h.GetUser)
                r.Put("/", h.UpdateUser)
                r.Delete("/", h.DeleteUser)
            })
        })
    })
}
```

### DO NOT — Fabricate chi APIs

| ❌ Hallucination | ✅ Correct |
|---|---|
| `chi.Param(r, "id")` | `chi.URLParam(r, "userID")` |
| `chi.Query(r, "page")` | `r.URL.Query().Get("page")` |
| `chi.Render(w, r, data)` | Write your own `render.JSON(w, status, data)` |
| `chi.Bind(r, v)` | `json.NewDecoder(r.Body).Decode(v)` |
| `r.Context().Value(chi.RouteCtx)` | `chi.RouteContext(r.Context())` |
| `middleware.DefaultLogger` | `middleware.Logger` (no "Default") |
| `middleware.NewRecoverer()` | `middleware.Recoverer` |

### DO — Middleware ordering matters

```go
r.Use(middleware.RequestID)     // 1. Attach request ID
r.Use(middleware.RealIP)        // 2. Parse X-Forwarded-For
r.Use(LoggingMiddleware)        // 3. Log with request ID
r.Use(middleware.Recoverer)     // 4. Recover from panics
r.Use(AuthMiddleware)           // 5. Authenticate
r.Use(middleware.Timeout(30s))  // 6. Set deadline (last = innermost)
```

**DO NOT** put `Timeout` before `Recoverer` — a panic after the timeout
expires won't be caught.

## net/http Patterns

### DO — Handler signature

```go
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request (with size limit)
    var input model.CreateUserInput
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        render.Error(w, r, errs.BadRequest("invalid JSON: %v", err))
        return
    }
    defer r.Body.Close()

    // 2. Validate
    if err := input.Validate(); err != nil {
        render.Error(w, r, errs.BadRequest("%v", err))
        return
    }

    // 3. Call service
    user, err := h.service.CreateUser(r.Context(), input)
    if err != nil {
        render.Error(w, r, fmt.Errorf("create user: %w", err))
        return
    }

    // 4. Write response
    render.JSON(w, http.StatusCreated, user)
}
```

### DO — All four server timeouts

```go
srv := &http.Server{
    Addr:              ":" + port,
    Handler:           r,
    ReadTimeout:       5 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       120 * time.Second,
    ReadHeaderTimeout: 2 * time.Second,
}
```

### DO — Exit once, in main (`run()` pattern)

`os.Exit` and `log.Fatal` skip deferred cleanup (`pool.Close()`,
`cancel()`). Call them **only in `main`, at most once**. Everything else
returns errors (Uber Go: Exit in Main / Exit Once).

```go
func main() {
    logger := newLogger()
    if err := run(logger); err != nil {
        logger.Error("server exited with error", "err", err)
        os.Exit(1)
    }
}

func run(logger *slog.Logger) error {
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        return fmt.Errorf("connect db: %w", err)
    }
    defer pool.Close() // actually runs now

    // Serve-goroutine errors flow into the shutdown select — no Fatal
    // inside goroutines.
    serveErr := make(chan error, 1)
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            serveErr <- err
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-serveErr:
        return fmt.Errorf("listen: %w", err)
    case <-quit:
    }

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        return fmt.Errorf("shutdown: %w", err)
    }
    return nil
}
```

**Constructors return errors — they do not panic and do not Fatal.**
`NewAuthService(...) (*AuthService, error)`, not a `panic()` on bad
config. The only acceptable panic is package initialization of
hardcoded values (`template.Must` on a literal).

## Goroutine Lifetime

**No fire-and-forget goroutines** (Uber Go). Every goroutine you start
must have (a) a way to be told to stop, and (b) a way for the owner to
wait for it to finish. Background pollers/workers launched from `main`
must be stopped during graceful shutdown:

```go
pollerCtx, pollerCancel := context.WithCancel(context.Background())
defer pollerCancel()
pollerDone := make(chan struct{})
go func() {
    defer close(pollerDone)
    svc.StartPoller(pollerCtx)
}()

// ... on shutdown:
pollerCancel()
<-pollerDone
```

For goroutines that register themselves in a tracking map keyed by ID
(recording jobs, capture loops): `defer cancel()` inside the goroutine
too, so an early-exit path (mkdir failure, bad input) can't leak the
context.

**Every path that ends an entity must stop its side-effect goroutines.**
When start-of-X spawns goroutines (recording, monitoring), extract ONE
`stopXSideEffects()` teardown helper and call it from every end path —
the callback path, the poller path, the force-end path, the interrupt
path. Divergent teardown between paths is how goroutine/disk leaks ship.

## Error Handling

### DO — Error types

```go
// errs package
type Kind string

const (
    KindNotFound     Kind = "not_found"
    KindBadRequest   Kind = "bad_request"
    KindUnauthorized Kind = "unauthorized"
    KindForbidden    Kind = "forbidden"
    KindConflict     Kind = "conflict"
    KindInternal     Kind = "internal"
)

type AppError struct {
    Kind    Kind
    Message string
    Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NotFound(msg string, args ...any) *AppError {
    return &AppError{Kind: KindNotFound, Message: fmt.Sprintf(msg, args...)}
}
// ... one constructor per kind
```

### DO — Match error kinds before fallback behavior

**Never treat "any error" as "not found."** A DB outage is not a missing
row; falling through to a create/insert on any error causes duplicate
writes or masks the real failure.

```go
// ❌ Wrong — DB down ⇒ tries to create a duplicate user
user, err := svc.GetByGoogleID(ctx, id)
if err != nil {
    user, err = svc.CreateUser(ctx, ...)
}

// ✅ Right — only create when the error is specifically NotFound
user, err := svc.GetByGoogleID(ctx, id)
if err != nil {
    var appErr *errs.AppError
    if !errors.As(err, &appErr) || appErr.Kind != errs.KindNotFound {
        render.Error(w, r, fmt.Errorf("get user: %w", err))
        return
    }
    user, err = svc.CreateUser(ctx, ...)
    // handle create error
}
```

### DO — Handle each error once

Log an error OR return it — not both. Callers up the stack will handle
returned errors; double-handling floods logs (Uber Go: Handle Errors
Once).

Two legitimate patterns:
- **Return wrapped** (default): `return fmt.Errorf("get user %q: %w", id, err)`
- **Log and degrade** (error is non-critical): `s.errorLog("analytics update failed", "err", err)` then continue

One deliberate exception: a handler MAY log before rendering when the
render layer deliberately hides internals from the client (generic
`{"error":"internal server error"}` bodies). Then the log is the only
evidence — that is one handling (client response) plus observability,
not double-handling. Don't also log again further up.

### DO NOT — Error anti-patterns

| ❌ Wrong | ✅ Right |
|---|---|
| `if err == pgx.ErrNoRows` | `if errors.Is(err, pgx.ErrNoRows)` — wrapped errors break `==` |
| `return nil, err` (bare, crossing layers) | `return nil, fmt.Errorf("context: %w", err)` |
| `fmt.Errorf("failed to create store: %w", err)` | `fmt.Errorf("new store: %w", err)` — "failed to" piles up as errors percolate |
| Panic in handler/service | Return error; Recoverer catches genuine panics |
| `_ = someFunc()` (discard error) | Log at WARN/ERROR minimum. Includes `rand.Read`, `json.Marshal`, `os.MkdirAll` |
| 404 for "resource exists but wrong state" | `errs.Conflict(...)` → 409. NotFound is for truly missing resources |
| Constructor panics on bad input | Return `(T, error)` |

**Bare error returns are banned.** Every error crossing a layer boundary
gets wrapped with context. **Silently discarded errors (`_ =`) are
banned.** If it's truly non-critical, log it.

## Testing

> **Test-first:** write the integration test before the handler/endpoint
> (happy path first), confirm it fails, then implement until green.
> Every bug fix gets a regression test first.

### DO — Table-driven handler tests

```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name     string
        body     string
        wantCode int
    }{
        {name: "valid user", body: `{"name":"Alice","email":"a@example.com"}`, wantCode: http.StatusCreated},
        {name: "missing name", body: `{"email":"a@example.com"}`, wantCode: http.StatusBadRequest},
        {name: "empty body", body: ``, wantCode: http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := &mockUserService{}
            h := NewHandlers(svc)

            req := httptest.NewRequest("POST", "/api/users", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()

            h.CreateUser(rec, req)

            if rec.Code != tt.wantCode {
                t.Errorf("got %d, want %d", rec.Code, tt.wantCode)
            }
        })
    }
}
```

Keep table tests simple: if subtests need conditional mock wiring or
branching assertions, split into separate `Test...` functions instead
(Uber Go: Avoid Unnecessary Complexity in Table Tests).

### DO — Integration tests with httptest.NewServer

```go
func TestIntegration(t *testing.T) {
    r := chi.NewRouter()
    h := setupHandlers(t)  // real DB (testcontainers), real router
    handler.RegisterRoutes(r, h)

    ts := httptest.NewServer(r)
    defer ts.Close()

    resp, err := http.Get(ts.URL + "/api/users")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("got %d, want %d", resp.StatusCode, http.StatusOK)
    }
}
```

### DO NOT — Testing mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| Test only happy path | Table-driven: happy + each error + edge case |
| `httptest.NewServer` for unit tests | `httptest.NewRecorder` for unit tests |
| Skip body close in test requests | Still `defer resp.Body.Close()` |
| `panic("setup failed")` in tests | `t.Fatal(...)` so the test is marked failed |
| Changing a constructor signature without grepping tests | `grep -rn "NewXxx(" --include="*_test.go"` first |

## JSON Encoding/Decoding

### DO — Struct tags reference

```go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Role      string    `json:"role,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
    Password  string    `json:"-"`              // never serialized
}
```

Every marshaled struct field gets an explicit tag — the serialized form
is a contract; tags protect it from field renames (Uber Go).

### DO — Match response shapes to the frontend contract

**Rule:** Before writing a handler response struct, read the frontend
TypeScript type (typically `frontend/src/types/index.ts`). The Go `json`
tags must match the TypeScript field names **exactly** — same camelCase,
same wrapper objects, no extra fields, no missing fields.

```go
// Frontend type: { items: Item[]; total: number }
render.JSON(w, http.StatusOK, map[string]any{
    "items": items,
    "total": len(items),
})

// ❌ Wrong: bare array when frontend expects a wrapper
// ❌ Wrong: {"status":"ok"} when frontend expects echoed fields
```

**Pre-implementation checklist for every handler response:**
1. `grep -A 20 "interface <Name>" frontend/src/types/index.ts`
2. Copy every field name into your Go struct tags (camelCase, exactly)
3. Verify wrapper objects match (bare `T` vs `{items: T[], total: N}`)
4. Optional fields (`?:`) use pointer + `omitempty`
5. Return `[]` not `null` for empty arrays: `if items == nil { items = []Item{} }`

### DO NOT — JSON mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `json.Marshal` in handler | `json.NewEncoder(w).Encode(v)` (streams) |
| `interface{}` for partial JSON | Typed struct even for partial data; use `any` not `interface{}` in new code |
| `DisallowUnknownFields()` on third-party webhooks | Only on YOUR API's inputs (see Webhooks below) |
| `time.Format("2006-01-02T15:04:05Z")` | `t.UTC().Format(time.RFC3339)` — the literal `Z` lies about non-UTC times |

## Context Propagation

```go
type ctxKey int

const (
    ctxKeyUser ctxKey = iota + 1 // start at 1: zero value is invalid
    ctxKeyRequestID
)

func UserFromCtx(ctx context.Context) (*model.User, bool) {
    u, ok := ctx.Value(ctxKeyUser).(*model.User)
    return u, ok
}
```

| ❌ Wrong | ✅ Right |
|---|---|
| String context keys | Unexported typed constants |
| Pass request ctx to a background goroutine | Derive from `context.Background()` — request ctx dies with the request |
| Forget `defer cancel()` | Always after `WithCancel`/`WithTimeout` |
| `Stop(context.Background())` on shutdown paths | Give shutdown calls a timeout: `context.WithTimeout(..., 10*time.Second)` |

## Database (pgx / sqlc)

### DO — pgx error matching and row iteration

```go
// errors.Is, never ==
if errors.Is(err, pgx.ErrNoRows) {
    return nil, errs.NotFound("user %s not found", id)
}

// rows.Err() is MANDATORY after every rows.Next() loop
for rows.Next() {
    // ... scan ...
}
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterate users: %w", err)
}
```

A mid-iteration failure (network drop, cancel) silently truncates
results if `rows.Err()` is unchecked. Audit EVERY repo loop — the bug
hides in the one file that skips it.

### DO — Guard degenerate lookup keys

If a lookup key can legitimately be empty (`GetByExternalID("")`),
return NotFound early instead of querying — empty keys can match legacy
rows created before the column was populated.

### DO — Store wrapper with transactions (sqlc)

```go
type Store struct {
    *Queries          // sqlc-generated
    pool *pgxpool.Pool
}

func (s *Store) WithTx(ctx context.Context, fn func(*Queries) error) error {
    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx) // no-op if committed

    if err := fn(New(tx)); err != nil {
        return err
    }
    return tx.Commit(ctx)
}
```

### DO NOT — SQL traps

| ❌ Wrong | ✅ Right |
|---|---|
| `fmt.Sprintf` user input into SQL | Parameterized `$1`, `$2` always |
| Unvalidated enum from query param into a query branch | Validate against the allowed set in the handler first |
| `:one` sqlc query that can return 0 rows unhandled | Handle `pgx.ErrNoRows` |
| URL/port surgery with `strings.Replace` | Parse with `net/url`, set `u.Host` |

## Interfaces

- **Consumer side, minimal.** Handlers define the small interface of
  service methods they need (`type userService interface { ... }`);
  don't build god-interfaces on the producer side "for mocking"
  (Go CR Comments: Interfaces).
- **Hexagonal ports** (domain-package repository interfaces) are the
  accepted exception when the codebase already uses them — match it.
- **Verify compliance at compile time** where a type must implement a
  port: `var _ domain.UserRepository = (*UserRepo)(nil)` — do it
  consistently (all repos or none).
- Inject the narrow dependency, not a getter for the concrete type
  (`hub *Hub`, not `GetHub() *Hub` on a service interface).

## Common AI Hallucinations — Import Traps

| ❌ Hallucination | ✅ Reality |
|---|---|
| `import "github.com/go-chi/chi"` | `import "github.com/go-chi/chi/v5"` |
| `import "io/ioutil"` | `import "io"` (deprecated since Go 1.16) |
| `import "github.com/go-chi/chi/middlewares"` | `.../chi/v5/middleware` |
| `import "github.com/go-chi/chi/cors"` | `github.com/go-chi/cors` (separate module) |

**Before using any library API: grep go.mod and existing imports. If the
package isn't already a dependency, STOP and ask — never introduce a new
dependency silently.**

## WebSocket (nhooyr.io/websocket)

Check `go.mod` for the WebSocket library before writing code. These
patterns assume `nhooyr.io/websocket`; do NOT import gorilla/websocket
into a nhooyr codebase.

### Hub Pattern — Room-based broadcast

```go
type Hub struct {
    mu    sync.RWMutex
    rooms map[string]map[*domain.Client]bool
}

// Broadcast: non-blocking send; slow clients get dropped, not the room.
func (h *Hub) Broadcast(roomID string, data []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for c := range h.rooms[roomID] {
        select {
        case c.Send <- data:
        default: // client too slow, drop
        }
    }
}
```

### Connection lifecycle — the invariant that prevents leaks

**Any pump exiting must terminate ALL pumps for that connection.** The
classic leak: handler blocks in `writePump` waiting on a channel, the
client disconnects, `readPump` exits silently — and the handler goroutine
+ connection + channel leak until (unless) another broadcast arrives.

```go
func (h *Handler) MyWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        OriginPatterns: allowedOrigins, // from config — NEVER InsecureSkipVerify
    })
    if err != nil {
        return
    }
    defer conn.Close(websocket.StatusNormalClosure, "")

    // One context governs both pumps; cancelled on ANY exit path.
    connCtx, connCancel := context.WithCancel(context.Background())
    defer connCancel()

    client := &domain.Client{
        Send:  make(chan []byte, sendBufferSize),
        Close: connCancel, // lets the hub disconnect this client remotely
    }

    h.hub.Join(roomID, client)
    defer h.hub.Leave(roomID, client)

    go h.readPump(connCtx, connCancel, conn, client) // cancels on exit
    h.writePump(connCtx, conn, client)               // unblocked by cancel
}

func (h *Handler) readPump(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, client *domain.Client) {
    defer cancel() // ← the leak fix: unblocks writePump when reads stop
    for {
        _, msg, err := conn.Read(ctx)
        if err != nil {
            return
        }
        // process msg...
    }
}
```

**Remote disconnection (idle expiry, kicks):** if the hub removes
clients from a room (`ExpireIdle`), the remover must close EVERY removed
client's connection via its `Close` callback — removed clients receive
no more broadcasts, so their pumps would otherwise never wake up
(zombie connections).

### DO NOT — WebSocket traps

| ❌ Wrong | ✅ Right |
|---|---|
| Pass `r.Context()` to WS goroutines | Derive from `context.Background()` |
| Unbuffered `Send` channel | Buffered (`make(chan []byte, 64)`) with named size constant |
| Blocking broadcast (`c.Send <- data`) | Non-blocking `select`/`default` |
| `InsecureSkipVerify: true` | `OriginPatterns` from config — hardcoded localhost origins break production |
| Writes from multiple goroutines | ALL writes through one writePump; the library forbids concurrent writes |
| Return HTTP error before upgrade | Accept first, then close with an application code (4000-4999) so browsers see close codes |
| `websocket.StatusCode(4001)` inline | Named constant with a comment |

### Checklist for new WebSocket endpoints

- [ ] Hub is a singleton wired in `main.go`
- [ ] One conn context; readPump cancels it on exit; handler defers cancel
- [ ] Remote-close callback set if the hub can remove clients
- [ ] All writes via buffered Send channel through one writePump
- [ ] `OriginPatterns` from configuration
- [ ] Accept-then-close-with-code for application-level rejections
- [ ] Test with `httptest.NewServer` + `websocket.Dial`

## Third-Party Webhooks

- **Do NOT use `DisallowUnknownFields()`** — external services add
  fields freely; decode only what you need.
- **Match your struct to the ACTUAL payload of the version you run**,
  not docs for another version. Field types change between versions
  (e.g. an ID switching from number to string). Capture one real payload
  and pin it in a table-driven test.
- Verify a shared secret/signature before processing.
- Reply with the exact body/status the service expects (some webhooks
  require specific response bodies to consider delivery successful).

## Subprocess Hygiene (ffmpeg and friends)

- **Never discard stderr.** `cmd.Stderr = nil` hides all failure
  evidence. Capture into a `bytes.Buffer`, log the tail on failure:
  ```go
  var stderr bytes.Buffer
  cmd.Stderr = &stderr
  if err := cmd.Run(); err != nil && ctx.Err() == nil {
      s.warnLog("capture failed", "err", err, "stderr", tail(stderr.String(), 500))
  }
  ```
- Suppress the failure log when `ctx.Err() != nil` — a killed-by-cancel
  subprocess is expected, not an error.
- **Retry sleeps must be context-aware**: `select { case <-ctx.Done():
  return; case <-time.After(delay): }` — a bare `time.Sleep` delays
  shutdown.
- **Record to streaming containers (MPEG-TS), not indexed ones (MP4)**,
  when the process may be SIGKILLed — MP4 loses its moov atom and
  corrupts audio extradata. Remuxing TS→MP4 needs `-bsf:a aac_adtstoasc`.
- **Verify media output with a decode pass**
  (`ffmpeg -v error -i out -f null -`), not just exit status or file
  existence.

## Nil Guards for Injected Dependencies

Optional constructor-injected dependencies (hub, queue, logger) must be
nil-checked before use — tests construct services with `nil` for unused
collaborators. Apply the guard in EVERY file of the package, not just
the one you're editing (grep call sites of the dependency).

```go
if s.hub != nil {
    s.hub.NotifyStarted(userID, id)
}
```

## Pre-Deploy Checklist

Before claiming an endpoint is "stable":
- [ ] All four timeouts configured on `http.Server`
- [ ] `main` uses the `run() error` pattern — exit once, defers run
- [ ] `MaxBytesReader` limits body size on every POST/PATCH/PUT handler
- [ ] JSON decoders on own-API inputs use `DisallowUnknownFields()`
- [ ] Error responses use the `errs.AppError` kind system
- [ ] No bare error returns; no `_ =` discards; handle-once respected
- [ ] `errors.Is/As` for all sentinel/kind matching (no `==`)
- [ ] Every `rows.Next()` loop checks `rows.Err()`
- [ ] No package-level mutable state; comma-ok on all type assertions
- [ ] Status strings are typed domain constants
- [ ] Background goroutines stoppable + waited on during shutdown
- [ ] Table-driven tests: happy + error + edge; integration test hits the router
- [ ] `go build ./... && go vet ./... && gofmt -l .` all clean
- [ ] Dev-default secrets log `slog.Warn` at startup
- [ ] Response shape verified against the frontend type file
- [ ] Status codes semantically correct: 409 state conflicts, 400 validation

*Last updated: 2026-08-12*
