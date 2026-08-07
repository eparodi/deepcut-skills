---
name: go-chi
description: Go/chi backend development standards — project layout, layering, error handling, testing, database patterns, and platform-specific AI traps. Load when writing Go code in the backend/ directory.
---

# Go/chi Backend Standards

You are writing Go code using net/http stdlib + chi router v5. Follow
these conventions exactly. When DeepSeek suggests something that
contradicts this document, trust this document.

## Project Layout (Monorepo)

```
backend/
├── cmd/
│   └── server/
│       └── main.go            # Entry point, wires everything
├── internal/
│   ├── handler/               # HTTP handlers (one file per resource)
│   │   ├── users.go
│   │   ├── products.go
│   │   └── routes.go          # chi route registration
│   ├── service/               # Business logic
│   │   ├── users.go
│   │   └── products.go
│   ├── store/                 # Data access (database, external APIs)
│   │   ├── postgres/
│   │   │   ├── queries/       # sqlc query files (.sql)
│   │   │   ├── db.go          # pgxpool setup
│   │   │   ├── models.go      # sqlc-generated (DO NOT EDIT)
│   │   │   └── querier.go     # sqlc-generated interface
│   │   │   ├── users.go       # Repository methods
│   │   │   └── products.go
│   │   └── memstore/          # In-memory store for tests
│   ├── middleware/             # Custom middleware
│   │   ├── auth.go
│   │   ├── logging.go
│   │   └── requestid.go
│   ├── model/                 # Domain types (shared across layers)
│   │   ├── user.go
│   │   └── product.go
│   ├── errs/                  # Custom error types
│   │   └── errors.go
│   └── config/                # Configuration
│       └── config.go
├── db/
│   └── migrations/            # golang-migrate SQL files
│       ├── 000001_create_users.up.sql
│       └── 000001_create_users.down.sql
├── go.mod
├── go.sum
└── Makefile
```

### Layer Rules

| Layer | Directory | Can import | Must NOT import |
|-------|-----------|-----------|-----------------|
| Handler | `internal/handler/` | `service`, `model`, `errs`, `middleware` | `store` directly |
| Service | `internal/service/` | `store`, `model`, `errs` | `handler`, `net/http` |
| Store | `internal/store/` | `model`, `errs` | `handler`, `service`, `net/http` |
| Model | `internal/model/` | nothing | `handler`, `service`, `store` |

**Rule:** Handlers parse HTTP, services enforce business rules, stores
talk to databases. A handler never opens a database connection. A service
never parses an HTTP request body.

## chi Router Patterns

### DO — Standard route registration

```go
// internal/handler/routes.go
package handler

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

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
| `chi.Render(w, r, data)` | Write your own `render.JSON(w, status, data)` |
| `chi.NewRouter()` returning `*chi.Mux` | `chi.NewRouter()` returns `chi.Router` |
| `r.Context().Value(chi.RouteCtx)` | Use `chi.RouteContext(r.Context())` |
| `middleware.DefaultLogger` | `middleware.Logger` (no "Default") |

### DO — Context middleware for entity loading

```go
func (h *Handlers) UserCtx(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := chi.URLParam(r, "userID")
        if userID == "" {
            render.Error(w, r, errs.BadRequest("missing user ID"))
            return
        }
        user, err := h.service.GetUser(r.Context(), userID)
        if err != nil {
            render.Error(w, r, fmt.Errorf("get user: %w", err))
            return
        }
        ctx := context.WithValue(r.Context(), ctxKeyUser, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### DO — Middleware ordering matters

```go
// Correct order (first defined = outermost, runs first)
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
    // 1. Parse request
    var input model.CreateUserInput
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

### DO NOT — Common handler mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| Forget `defer r.Body.Close()` | Always `defer r.Body.Close()` after reading |
| Set headers after `WriteHeader` | `Header().Set()` BEFORE `WriteHeader()` |
| Write response before checking error | Check error FIRST, then write response |
| `w.Write([]byte(...))` for JSON | Use `json.NewEncoder(w).Encode(v)` |
| Ignore `r.Body` size | Use `http.MaxBytesReader` to limit body size |

### DO — All four server timeouts

```go
srv := &http.Server{
    Addr:         ":" + port,
    Handler:      r,
    ReadTimeout:  5 * time.Second,   // time to read request
    WriteTimeout: 10 * time.Second,  // time to write response
    IdleTimeout:  120 * time.Second, // keep-alive timeout
    ReadHeaderTimeout: 2 * time.Second, // time to read headers only
}
```

### DO — Graceful shutdown

```go
func main() {
    // ... setup ...

    srv := &http.Server{Addr: ":8080", Handler: r}

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("shutdown: %v", err)
    }
}
```

## Error Handling

### DO — Error types

```go
// internal/errs/errors.go
package errs

import "fmt"

type Kind string

const (
    KindNotFound       Kind = "not_found"
    KindBadRequest     Kind = "bad_request"
    KindUnauthorized   Kind = "unauthorized"
    KindForbidden      Kind = "forbidden"
    KindConflict       Kind = "conflict"
    KindInternal       Kind = "internal"
)

type AppError struct {
    Kind    Kind
    Message string
    Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// Constructor helpers
func NotFound(msg string, args ...any) *AppError {
    return &AppError{Kind: KindNotFound, Message: fmt.Sprintf(msg, args...)}
}
func BadRequest(msg string, args ...any) *AppError {
    return &AppError{Kind: KindBadRequest, Message: fmt.Sprintf(msg, args...)}
}
// ... etc for each kind
```

### DO NOT — Sentinel errors as strings

```go
// ❌ Bad: string comparison
var ErrNotFound = errors.New("not found")
if errors.Is(err, ErrNotFound) { ... } // fragile

// ✅ Good: type assertion
var appErr *errs.AppError
if errors.As(err, &appErr) {
    switch appErr.Kind {
    case errs.KindNotFound:
        w.WriteHeader(http.StatusNotFound)
    }
}
```

### DO NOT — Error handling anti-patterns

| ❌ Wrong | ✅ Right |
|---|---|
| `if err != nil { log.Fatal(err) }` in handlers | Return error to caller, let middleware log |
| `return nil, err` (bare) | `return nil, fmt.Errorf("context: %w", err)` |
| Panic in handler | Return error; Recoverer middleware catches panics |
| `errors.New("user not found")` | `errs.NotFound("user %s not found", id)` |

## Testing

### DO — Table-driven handler tests

```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        body    string
        wantCode int
        wantErr  bool
    }{
        {
            name:     "valid user",
            body:     `{"name":"Alice","email":"alice@example.com"}`,
            wantCode: http.StatusCreated,
        },
        {
            name:     "missing name",
            body:     `{"email":"alice@example.com"}`,
            wantCode: http.StatusBadRequest,
        },
        {
            name:     "empty body",
            body:     ``,
            wantCode: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := &mockUserService{}
            h := NewHandlers(svc)

            req := httptest.NewRequest("POST", "/api/users",
                strings.NewReader(tt.body))
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

### DO — Integration tests with httptest.NewServer

```go
func TestIntegration(t *testing.T) {
    // Setup real router with real dependencies
    r := chi.NewRouter()
    h := setupHandlers(t)  // real DB connection, etc.
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
| Mock database by hand | Use generated mocks (mockgen) or interface stubs |
| `httptest.NewServer` for unit tests | `httptest.NewRecorder` for unit tests |
| Skip body close in test requests | Still `defer resp.Body.Close()` |

## JSON Encoding/Decoding

### DO — Struct tags reference

```go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Role      string    `json:"role,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
    Password  string    `json:"-"`              // never serialized
    Internal  string    `json:"-,"`             // literal "-" key
    Count     int       `json:"count,string"`   // "42" instead of 42
}
```

### DO — Disallow unknown fields

```go
decoder := json.NewDecoder(r.Body)
decoder.DisallowUnknownFields()
var input model.CreateUserInput
if err := decoder.Decode(&input); err != nil {
    // handle error
}
```

### DO NOT — JSON mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `json.Marshal` in handler | `json.NewEncoder(w).Encode(v)` (streams) |
| Ignore JSON syntax errors | Switch on `*json.SyntaxError`, `*json.UnmarshalTypeError` |
| `interface{}` for partial JSON | Define a typed struct even for partial data |

## Context Propagation

### DO — Typed, unexported context keys

```go
// internal/handler/context.go
package handler

type ctxKey int

const (
    ctxKeyUser ctxKey = iota
    ctxKeyRequestID
)

func UserFromCtx(ctx context.Context) (*model.User, bool) {
    u, ok := ctx.Value(ctxKeyUser).(*model.User)
    return u, ok
}

func WithUser(ctx context.Context, u *model.User) context.Context {
    return context.WithValue(ctx, ctxKeyUser, u)
}
```

### DO NOT — Context mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| String context keys | Unexported typed constants |
| Pass request ctx to goroutine | `ctx, cancel := context.WithTimeout(context.Background(), ...)` |
| Forget `defer cancel()` | Always `defer cancel()` after `WithCancel`/`WithTimeout` |

## Database (sqlc + pgx)

### DO — sqlc configuration

```yaml
# db/sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/store/postgres/queries/"
    schema: "db/migrations/"
    gen:
      go:
        package: "postgres"
        out: "internal/store/postgres/"
        sql_package: "pgx/v5"
        emit_interface: true
        emit_json_tags: true
```

### DO — sqlc annotations

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateUser :exec
INSERT INTO users (name, email) VALUES ($1, $2);

-- name: UpdateUser :execrows
UPDATE users SET name = $2, email = $3 WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

### DO NOT — sqlc traps

| ❌ Wrong | ✅ Right |
|---|---|
| `:one` for query that might return 0 rows | `:one` returns `sql.ErrNoRows`; handle it |
| Raw `database/sql` calls | Always go through sqlc-generated `Queries` |
| `:exec` when you need affected rows | Use `:execrows` to get `sql.Result.RowsAffected()` |
| Manual SQL in Go strings | Queries in `.sql` files; run `make sqlc` to regenerate |

### DO — Store wrapper with transactions

```go
// internal/store/postgres/db.go
type Store struct {
    *Queries           // embeds sqlc-generated interface
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
    return &Store{
        Queries: New(pool),
        pool:    pool,
    }
}

func (s *Store) WithTx(ctx context.Context, fn func(*Queries) error) error {
    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx) // no-op if committed

    q := New(tx)
    if err := fn(q); err != nil {
        return err
    }
    return tx.Commit(ctx)
}
```

## Common AI Hallucinations — Complete Reference

### chi API Fabrications

| ❌ DeepSeek says | ✅ Reality |
|---|---|
| `chi.Param(r, "id")` | `chi.URLParam(r, "id")` |
| `chi.Query(r, "page")` | `r.URL.Query().Get("page")` |
| `chi.Render(w, r, v)` | Write your own `render.JSON(w, code, v)` |
| `chi.Bind(r, v)` | Use `json.NewDecoder(r.Body).Decode(v)` |
| `chi.NewRouter()` typed as `*chi.Mux` | Returns `chi.Router` |
| `chi.Get(r.Context())` | `chi.RouteContext(r.Context())` |
| `middleware.DefaultLogger` | `middleware.Logger` |
| `middleware.NewRecoverer()` | `middleware.Recoverer` |
| `middleware.DefaultCompress` | `middleware.Compress(5, "application/json")` |

### Import Traps

| ❌ DeepSeek says | ✅ Reality |
|---|---|
| `import "github.com/go-chi/chi"` | `import "github.com/go-chi/chi/v5"` |
| `import "io/ioutil"` | `import "io"` (ioutil deprecated in Go 1.16) |
| `import "github.com/go-chi/chi/middlewares"` | `import "github.com/go-chi/chi/v5/middleware"` |
| `import "github.com/go-chi/chi/cors"` | `import "github.com/go-chi/cors"` (separate module) |

### Goroutine / Handler Traps

| ❌ Wrong | ✅ Right |
|---|---|
| Launch goroutine with `r.Context()` | Derive new context: `context.Background()` or `context.WithTimeout(context.Background(), ...)` |
| No panic recovery in goroutine | `defer func() { if r := recover(); r != nil { log.Error(...) } }()` |
| `w.WriteHeader(200)` after `w.Write()` | Set status code BEFORE writing body |
| Return 200 then write error body | Check errors FIRST, then write success OR error |

### Pre-Deploy Checklist

Before claiming an endpoint is "stable":
- [ ] All four timeouts configured on `http.Server`
- [ ] `MaxBytesReader` limits request body size
- [ ] Every handler has `defer r.Body.Close()`
- [ ] Error responses use the `errs.AppError` type system
- [ ] Table-driven tests cover: happy path + each error path + edge cases
- [ ] Integration test hits the router with a real request
- [ ] `go build ./...` and `go vet ./...` pass
