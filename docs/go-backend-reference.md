# Go Backend Reference: chi v5 + net/http — Traps, Mistakes & Best Practices

> **Audience:** AI coding agents (DeepSeek) and developers writing production Go HTTP services.
> **Versions:** Go 1.22+, chi v5.2.x, pgx v5.x, sqlc v1.28.x

---

## Table of Contents

1. [Monorepo Project Layout](#1-monorepo-project-layout)
2. [chi Router: Route Grouping, Middleware, Context](#2-chi-router)
3. [net/http: Handlers, ResponseWriter, Request Body, Timeouts, Graceful Shutdown](#3-nethttp)
4. [Error Handling](#4-error-handling)
5. [Testing](#5-testing)
6. [JSON Encoding/Decoding](#6-json)
7. [Context Propagation](#7-context-propagation)
8. [Middleware Patterns](#8-middleware-patterns)
9. [Database Integration (sqlc + pgx)](#9-database)
10. [Common AI Hallucinations & Mistakes](#10-ai-hallucinations)

---

## 1. Monorepo Project Layout

### DO: Standard Go monorepo layout for a backend service

```
project-root/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # single binary entrypoint
│   ├── internal/                     # Go's visibility boundary — unimportable externally
│   │   ├── handler/                  # HTTP handlers (thin: parse, call service, write response)
│   │   │   ├── user.go
│   │   │   └── user_test.go
│   │   ├── service/                  # business logic layer
│   │   │   ├── user.go
│   │   │   └── user_test.go
│   │   ├── store/                    # data access (sqlc queries + pgx pool wrapper)
│   │   │   ├── db/
│   │   │   │   ├── query.sql        # sqlc input
│   │   │   │   ├── query.sql.go     # sqlc generated
│   │   │   │   ├── models.go        # sqlc generated
│   │   │   │   └── db.go           # sqlc generated
│   │   │   ├── migrations/
│   │   │   │   ├── 001_create_users.up.sql
│   │   │   │   └── 001_create_users.down.sql
│   │   │   └── pg.go                # pgx pool constructor
│   │   ├── middleware/
│   │   │   ├── logging.go
│   │   │   ├── auth.go
│   │   │   ├── requestid.go
│   │   │   └── recovery.go
│   │   ├── model/                    # shared domain types (no imports from handler/service/store)
│   │   │   └── user.go
│   │   ├── errs/                     # sentinel errors + custom error types
│   │   │   └── errs.go
│   │   └── config/
│   │       └── config.go            # env parsing, validation
│   ├── go.mod
│   └── go.sum
├── frontend/                          # if applicable, entirely separate module
└── Makefile                           # task runner (generate, test, lint, migrate)
```

### DO NOT: Common layout mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `backend/main.go` at root; no `cmd/` | `backend/cmd/server/main.go` — allows multiple binaries |
| `backend/pkg/` for internal code | `backend/internal/` — enforces visibility, prevents external imports |
| `handler/` calls database directly | `handler → service → store` — layered, testable |
| Generated code mixed with hand-written | `store/db/` holds only generated code; `store/pg.go` is hand-written |
| `model/` imports `store/` or `handler/` | `model/` is a leaf package — zero internal dependencies |

### DO: `cmd/server/main.go` skeleton

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      newRouter(),           // chi router
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
```

> **Trap:** DeepSeek often forgets `os.Signal` is the package name AND the type name. The correct import is `"os/signal"` and the function is `signal.NotifyContext(...)`.

---

## 2. chi Router

### Route Grouping & Sub-routers

```go
import "github.com/go-chi/chi/v5"

func newRouter() chi.Router {
	r := chi.NewRouter()

	// --- Global middleware (applied to ALL routes) ---
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// --- Public routes ---
	r.Group(func(r chi.Router) {
		r.Post("/api/v1/login", handler.Login)
		r.Post("/api/v1/register", handler.Register)
	})

	// --- Authenticated routes ---
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)            // applies only within this group

		r.Route("/api/v1/users", func(r chi.Router) {
			r.Get("/", handler.ListUsers)
			r.Post("/", handler.CreateUser)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(userCtxMiddleware) // loads user into context via URL param

				r.Get("/", handler.GetUser)
				r.Put("/", handler.UpdateUser)
				r.Delete("/", handler.DeleteUser)
			})
		})
	})

	return r
}
```

### Middleware Ordering — THIS ORDER MATTERS

```
Request →
  1. RequestID      (set first — all subsequent middleware + handlers see it)
  2. RealIP         (parse X-Forwarded-For before logging)
  3. Logger         (log with request ID and real IP)
  4. Recoverer      (recover from panics AFTER logging so panics are logged)
  5. Timeout        (outermost deadline; chi v5 built-in)
  ... route-specific middleware ...
  ... handler ...
```

| ❌ DO NOT | ✅ DO |
|-----------|-------|
| Place `Recoverer` before ` Logger` — logged request data missing in panic traces | `RequestID → RealIP → Logger → Recoverer → Timeout` |
| Place `Timeout` after heavy middleware — auth, DB calls burn deadline | Wrap entire router in `Timeout` via `r.Use()` |
| Apply auth middleware to `r` (global) when you have public login | Use `r.Group()` for authenticated routes only |

### URL Parameters

```go
// ✅ Correct: chi.URLParam
func GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")   // from {userID} in pattern
	// ...
}

// ❌ Wrong: chi does NOT have chi.Param, chi.Parameter, chi.URLParam(r, "id") on a context
// ❌ Wrong: do not extract from r.Context().Value("userID")
```

**Trap:** The URL param key must match exactly the `{name}` in the route pattern. `chi.URLParam(r, "userId")` on `/{userID}` returns `""` with no error — a silent bug.

### Context Middleware (Loading Entities by URL Param)

```go
// ✅ DO: encapsulate context key in its own package/type
type contextKey string

const userContextKey contextKey = "user"

func UserCtxMiddleware(userService *service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := chi.URLParam(r, "userID")

			user, err := userService.GetByID(r.Context(), userID)
			if err != nil {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(userContextKey).(*model.User)
	return user, ok
}
```

### chi Built-in Middleware

| Import | Middleware | Notes |
|--------|-----------|-------|
| `"github.com/go-chi/chi/v5/middleware"` | `RequestID` | Injects unique per-request ID; reads from `X-Request-ID` header if present |
| `"github.com/go-chi/chi/v5/middleware"` | `Logger` | Structured request logging |
| `"github.com/go-chi/chi/v5/middleware"` | `Recoverer` | Recovers panics, logs stack trace, returns 500 |
| `"github.com/go-chi/chi/v5/middleware"` | `RealIP` | Parses `X-Forwarded-For` / `X-Real-IP` |
| `"github.com/go-chi/chi/v5/middleware"` | `Timeout` | Per-request deadline context |
| `"github.com/go-chi/chi/v5/middleware"` | `Heartbeat` | Responds 200 to `GET /health` (no DB check) |
| `"github.com/go-chi/chi/v5/middleware"` | `CleanPath` | Double slashes → single slash |
| `"github.com/go-chi/cors"` | `cors.Handler` | Separate module, NOT in chi core |

**Trap:** `middleware.Recoverer` uses `debug.Stack()` for the trace. In production you may want a custom one that uses `runtime.Stack()` with `all=false`.

### Route-Not-Found Handling

```go
r.NotFound(func(w http.ResponseWriter, r *http.Request) {
	// chi provides r.URL.Path and r.Method
	encodeJSON(w, http.StatusNotFound, map[string]string{
		"error":  "route not found",
		"path":   r.URL.Path,
		"method": r.Method,
	})
})

r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
	encodeJSON(w, http.StatusMethodNotAllowed, map[string]string{
		"error": "method not allowed",
	})
})
```

---

## 3. net/http

### Handler Signatures

```go
// ✅ Standard library handler: always this signature
func MyHandler(w http.ResponseWriter, r *http.Request) { ... }

// ✅ chi returns a Handler interface, not a HandlerFunc
var _ http.Handler = chi.NewRouter()   // true

// ❌ DO NOT: chi.NewRouter() returns *chi.Mux, but you should type it as chi.Router
// ✅ DO: use chi.Router interface
func newRouter() chi.Router { ... }
```

### ResponseWriter Rules

```go
// ✅ DO: check for http.Flusher before streaming
func streamHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	// ... write chunks, call flusher.Flush() ...
}

// ❌ WRONG: WriteHeader after Write
w.Write([]byte(`{"status":"ok"}`))
w.WriteHeader(http.StatusOK)  // NO — Write implicitly calls WriteHeader(200)

// ❌ WRONG: multiple WriteHeader calls
w.WriteHeader(http.StatusOK)
w.WriteHeader(http.StatusNotFound) // PANICS: http: superfluous response.WriteHeader call
```

**Ordering rule:** `Header().Set()` → `WriteHeader()` → `Write()`. Calling `Write` before `WriteHeader` sends a 200. Calling `WriteHeader` twice panics.

### Reading Request Body — Every Time

```go
// ✅ CORRECT: ALWAYS close the body, even on error paths
func CreateThing(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var input model.CreateThingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	// ...
}

// ❌ WRONG: forgetting to close — connection leaks under load
// ❌ WRONG: closing r.Body before Decode
// ❌ WRONG: using ioutil.ReadAll (deprecated since Go 1.16; use io.ReadAll)
// ✅ DO: limit body size
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

// ✅ DO: drain + close body to enable HTTP keep-alive
// (net/http does this automatically in most cases, but if you 
//  do NOT read the body, you must drain it)
```

### Server Timeouts (Critical for Production)

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           router,
	ReadTimeout:       5 * time.Second,    // time to read full request (including body)
	ReadHeaderTimeout: 2 * time.Second,    // time to read headers ONLY (protects against slow-loris)
	WriteTimeout:      10 * time.Second,   // time to write response
	IdleTimeout:       120 * time.Second,  // time to keep idle connections alive
	MaxHeaderBytes:    1 << 20,            // 1 MB max request header
}
```

**Trap:** `ReadTimeout` includes the body. If you expect large uploads, set `ReadHeaderTimeout` instead and leave `ReadTimeout = 0`, then enforce body-reading deadlines in the handler via `http.MaxBytesReader` + context deadline.

### Graceful Shutdown

```go
// ✅ Correct: signal.NotifyContext (Go 1.16+)
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

<-ctx.Done()

shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
defer shutdownCancel()

// Drain active connections
if err := srv.Shutdown(shutdownCtx); err != nil {
	slog.Error("shutdown incomplete", "err", err)
}

// ❌ DO NOT: srv.Close() — this immediately closes all connections, no draining
// ❌ DO NOT: log.Fatal on ListenAndServe when it returns http.ErrServerClosed
```

---

## 4. Error Handling

### Custom Error Type

```go
package errs

import "fmt"

// AppError is a structured error for HTTP responses.
type AppError struct {
	Code    int    `json:"-"`       // HTTP status code
	Kind    string `json:"kind"`    // machine-readable error kind, e.g. "not_found"
	Message string `json:"message"` // human-readable message
	Err     error  `json:"-"`       // wrapped underlying error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// --- Constructor helpers ---

func New(code int, kind, message string) *AppError {
	return &AppError{Code: code, Kind: kind, Message: message}
}

func Wrap(err error, code int, kind, message string) *AppError {
	return &AppError{Code: code, Kind: kind, Message: message, Err: err}
}
```

### Sentinel Errors

```go
package errs

import "errors"

var (
	ErrNotFound      = &AppError{Code: 404, Kind: "not_found", Message: "resource not found"}
	ErrUnauthorized  = &AppError{Code: 401, Kind: "unauthorized", Message: "authentication required"}
	ErrForbidden     = &AppError{Code: 403, Kind: "forbidden", Message: "access denied"}
	ErrValidation    = &AppError{Code: 422, Kind: "validation_error", Message: "request validation failed"}
	ErrInternal      = &AppError{Code: 500, Kind: "internal_error", Message: "internal server error"}
)

// ❌ DO NOT: use string sentinels — not comparable with errors.Is
// var ErrNotFound = "not found"   // WRONG

// ✅ Use errors.Is / errors.As in callers:
func GetUser(w http.ResponseWriter, r *http.Request) {
	user, err := userService.GetByID(r.Context(), id)
	if err != nil {
		var appErr *errs.AppError
		if errors.As(err, &appErr) {
			encodeJSON(w, appErr.Code, appErr)
			return
		}
		// Unknown error -> generic 500
		slog.Error("unhandled error", "err", err)
		encodeJSON(w, http.StatusInternalServerError, errs.ErrInternal)
		return
	}
	// ...
}
```

### Error Middleware Pattern

```go
// ✅ DO: define response envelope
type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Kind  string `json:"kind,omitempty"`
}

// ✅ DO: centralized error writing helper
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *errs.AppError
	if !errors.As(err, &appErr) {
		appErr = errs.ErrInternal
	}

	slog.ErrorContext(r.Context(), "request error",
		"method", r.Method,
		"path", r.URL.Path,
		"status", appErr.Code,
		"kind", appErr.Kind,
		"err", appErr.Error(),
	)

	encodeJSON(w, appErr.Code, Response{
		Error: appErr.Message,
		Kind:  appErr.Kind,
	})
}
```

### Wrapping Errors — Keep the Chain

```go
// ✅ DO: use fmt.Errorf with %w
return fmt.Errorf("userService.GetByID: %w", err)

// ❌ DO NOT: discard the original error
return fmt.Errorf("user not found") // loses all context

// ❌ DO NOT: log AND return error — the caller decides
// ❌ DO NOT: use %v — breaks errors.Is / errors.As
```

---

## 5. Testing

### Table-Driven Handler Tests

```go
func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func(*MockUserService)
		wantStatusCode int
		wantContains   string
	}{
		{
			name:           "valid user",
			body:           `{"name":"Alice","email":"alice@example.com"}`,
			mockSetup:      func(m *MockUserService) { m.CreateReturns(&model.User{ID: "1"}, nil) },
			wantStatusCode: http.StatusCreated,
			wantContains:   `"id":"1"`,
		},
		{
			name:           "invalid JSON",
			body:           `{bad`,
			mockSetup:      func(m *MockUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantContains:   "invalid",
		},
		{
			name:           "service error",
			body:           `{"name":"Bob","email":"bob@example.com"}`,
			mockSetup:      func(m *MockUserService) { m.CreateReturns(nil, errs.ErrInternal) },
			wantStatusCode: http.StatusInternalServerError,
			wantContains:   "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockSvc := &MockUserService{}
			tt.mockSetup(mockSvc)

			h := handler.NewUserHandler(mockSvc)
			r := chi.NewRouter()
			r.Post("/api/v1/users", h.Create)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("status = %d; want %d", rec.Code, tt.wantStatusCode)
			}
			if tt.wantContains != "" && !strings.Contains(rec.Body.String(), tt.wantContains) {
				t.Errorf("body does not contain %q; got %s", tt.wantContains, rec.Body.String())
			}
		})
	}
}
```

### chi Route Matching Tests

```go
func TestUserRouter(t *testing.T) {
	h := handler.NewUserHandler(mockService)

	tests := []struct {
		method string
		path   string
		status int // 0 = expect route match (status comes from handler)
		noRoute bool
	}{
		{method: "GET", path: "/api/v1/users", noRoute: false},
		{method: "POST", path: "/api/v1/users", noRoute: false},
		{method: "GET", path: "/api/v1/users/abc-123", noRoute: false},
		{method: "PUT", path: "/api/v1/users/abc-123", noRoute: false},
		{method: "DELETE", path: "/api/v1/users", noRoute: true},   // no delete on collection
		{method: "PATCH", path: "/api/v1/users/abc", noRoute: true}, // PATCH not registered
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			r := newRouter(h) // your constructor
			r.ServeHTTP(rec, req)

			if tt.noRoute && rec.Code != http.StatusMethodNotAllowed && rec.Code != http.StatusNotFound {
				t.Errorf("expected route to not match, got %d", rec.Code)
			}
		})
	}
}
```

### Integration Test with httptest.Server

```go
func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// ✅ DO: use real DB via testcontainers or a test database
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	r := NewRouter(pool)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// ✅ DO: use ts.URL to build full URLs
	resp, err := ts.Client().Post(
		ts.URL+"/api/v1/users",
		"application/json",
		strings.NewReader(`{"name":"integration"}`),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.Code)
}
```

### Test Fixtures

```go
// ✅ DO: store JSON fixtures as embedded files in testdata/
// backend/internal/handler/testdata/create_user_valid.json

//go:embed testdata/create_user_valid.json
var createUserValid string

// In the test:
req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(createUserValid))
```

### Mock Generation

Prefer `mockgen` (from `go.uber.org/mock`) for interface-based mocking:

```go
//go:generate mockgen -destination=../mocks/mock_user_service.go -package=mocks . UserService
type UserService interface {
	Create(ctx context.Context, input model.CreateUserInput) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}
```

---

## 6. JSON Encoding/Decoding

### Struct Tags

```go
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role,omitempty"`        // omit if empty string
	Score     *int      `json:"score,omitempty"`        // omit if nil pointer — use pointer for 0 vs absent
	CreatedAt time.Time `json:"created_at"`             // time.RFC3339 by default
	internal  string    `json:"-"`                      // always omitted
	Password  string    `json:"password,omitempty"`     // omit on marshaling, read on unmarshaling
	Unused    string    `json:"-"`                      // never marshaled or unmarshaled
}
```

### DO NOT: Common JSON mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `json.Decode(r.Body)` without `defer r.Body.Close()` | Always close |
| `json.NewDecoder(r.Body).Decode(&v)` twice on same body | Body is a read-once stream |
| Disallow unknown fields by default | Use `decoder.DisallowUnknownFields()` to catch typos |
| Ignore decoder errors beyond `io.EOF` | Check for `*json.SyntaxError`, `*json.UnmarshalTypeError` |
| Use `json.NewDecoder(r.Body)` when you need the raw bytes for logging | Decode from `io.ReadAll` bytes, then log raw before unmarshaling |

### Decoding Best Practices

```go
func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // ✅ reject unknown JSON keys

	if err := dec.Decode(v); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return errs.Wrap(err, http.StatusBadRequest, "bad_request",
				fmt.Sprintf("malformed JSON at position %d", syntaxError.Offset))
		case errors.As(err, &unmarshalTypeError):
			return errs.Wrap(err, http.StatusBadRequest, "bad_request",
				fmt.Sprintf("invalid type for field %q (expected %s)", 
					unmarshalTypeError.Field, unmarshalTypeError.Type))
		case errors.Is(err, io.EOF):
			return errs.New(http.StatusBadRequest, "bad_request", "request body is empty")
		default:
			return errs.Wrap(err, http.StatusBadRequest, "bad_request", "invalid JSON")
		}
	}
	return nil
}
```

### Streaming JSON Responses

```go
// ✅ DO: stream large responses
func ListLargeDataset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	// ✅ DO: flush after each record
	flusher, _ := w.(http.Flusher)

	for _, item := range items {
		if err := encoder.Encode(item); err != nil {
			// Client disconnected — stop
			return
		}
		flusher.Flush()
	}
}

// ❌ DO NOT: json.Marshal all items into []byte then w.Write — OOM on large datasets
```

### Custom Time Format

```go
type CustomTime struct{ time.Time }

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Format("2006-01-02T15:04:05.000Z07:00"))
}

// ❌ DO NOT: use time.Unix() integer for JSON — loses precision, not human-readable
```

---

## 7. Context Propagation

### Request-Scoped Values

```go
// ✅ DO: typed context keys (unexported, zero-size type)
type contextKey string

const (
	requestIDKey contextKey = "requestID"
	userKey      contextKey = "user"
)

// ✅ DO: typed getter/setter functions
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// ❌ DO NOT: use string literals as keys — collisions across packages
// ❌ DO NOT: use int, bool, or any builtin type as key — Go doesn't enforce uniqueness
// ❌ DO NOT: export context keys — other packages can overwrite
```

### Deadlines & Cancellation

```go
// ✅ DO: pass context through every layer, respect cancellation
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	// Check: is the request already cancelled?
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ✅ Pass ctx to database calls — pgx supports context cancellation
	row := s.pool.QueryRow(ctx, "SELECT ...", id)
	// ...
}

// ✅ DO: create sub-deadlines for known slow operations
func (s *UserService) GenerateReport(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel() // ✅ always call cancel to release timer resources
	// ...
}

// ❌ DO NOT: store context in a struct — it's request-scoped
// type Handler struct { ctx context.Context } // WRONG
// ❌ DO NOT: context.Background() deep in the call chain — use the caller's context
// ❌ DO NOT: forget to call cancel() — timer leak
```

### Context in Concurrent Code

```go
// ✅ DO: derive context for each goroutine
func ProcessBatch(ctx context.Context, ids []string) []Result {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan Result, len(ids))

	for _, id := range ids {
		go func(id string) {
			res, err := doWork(ctx, id)
			if err != nil {
				cancel() // ✅ cancel all siblings on first error
				return
			}
			results <- res
		}(id)
	}
	// ...
}

// ❌ DO NOT: use the request's context in a goroutine that outlives the request
// ❌ DO NOT: start a goroutine in a handler without a context.Context
```

---

## 8. Middleware Patterns

### RequestID Middleware

```go
// ✅ Use chi's built-in: middleware.RequestID

// ✅ To read it in your code:
import "github.com/go-chi/chi/v5/middleware"

func MyHandler(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	slog.Info("handling request", "requestID", reqID)
}

// ❌ DO NOT: implement your own — chi's handles X-Request-ID header, generation, and response headers
```

### Structured Logging Middleware

```go
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// ✅ Wrap ResponseWriter to capture status code
			wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			logger.InfoContext(r.Context(), "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration", time.Since(start).String(),
				"bytes", wrapped.bytes,
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK // if Write before WriteHeader
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

// ❌ DO NOT: log request body by default — could contain PII, passwords, tokens
// ❌ DO NOT: use log.Println in production — use structured logging
```

### Auth Middleware (Bearer Token)

```go
func AuthMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeError(w, r, errs.ErrUnauthorized)
				return
			}

			// ✅ DO: trim the Bearer prefix using strings.TrimPrefix
			token := strings.TrimPrefix(header, "Bearer ")
			if token == header { // didn't have Bearer prefix
				writeError(w, r, errs.ErrUnauthorized)
				return
			}

			claims, err := validateToken(token, jwtSecret)
			if err != nil {
				writeError(w, r, errs.Wrap(err, 401, "unauthorized", "invalid token"))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ❌ DO NOT: use strings.Split(header, "Bearer ") — edge cases with multiple "Bearer"
// ❌ DO NOT: validate token but not set context — downstream can't use it
```

### Recovery Middleware (Custom)

```go
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				err, ok := rec.(error)
				if !ok {
					err = fmt.Errorf("%v", rec)
				}

				// ✅ DO: get full stack trace
				buf := make([]byte, 4<<10) // 4 KB
				stackLen := runtime.Stack(buf, false) // false = only current goroutine

				logger.ErrorContext(r.Context(), "panic recovered",
					"err", err,
					"stack", string(buf[:stackLen]),
				)

				// ✅ DO: write structured error, not the stack
				encodeJSON(w, http.StatusInternalServerError, Response{
					Error: "internal server error",
					Kind:  "internal_error",
				})
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ❌ DO NOT: expose raw panic message to client — security leak
// ❌ DO NOT: use debug.Stack() in production — allocates entire stack of all goroutines
```

### CORS

```go
import "github.com/go-chi/cors"

func CORSConfig() cors.Options {
	return cors.Options{
		AllowedOrigins:   []string{"https://example.com"}, // ❌ NEVER: "*" with AllowCredentials
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// Usage:
r.Use(cors.Handler(corsConfig))

// ❌ DO NOT: use "*" origin with AllowCredentials: true — browser rejects it
// ❌ DO NOT: implement CORS manually — use go-chi/cors
```

---

## 9. Database Integration (sqlc + pgx, No ORM)

### sqlc Configuration

```yaml
# backend/sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/store/db/query.sql"
    schema: "internal/store/migrations/"
    gen:
      go:
        package: "db"
        out: "internal/store/db"
        sql_package: "pgx/v5"            # ✅ use pgx driver
        emit_json_tags: true              # ✅ auto add `json:"..."` struct tags
        emit_pointers_for_null_types: true # ✅ *string for nullable columns
        emit_params_struct_pointers: false
        query_parameter_limit: 3          # generates cleaner code
```

**Trap:** sqlc needs a _single_ `query.sql` (or multiple files in the queries dir). If you split queries across files, list them all or use a directory.

### sqlc Query File

```sql
-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, name, email, created_at, updated_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :exec
UPDATE users
SET name = $2, email = $3, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = $1;
```

| Annotation | Generated Return |
|------------|-----------------|
| `:one` | `(Row, error)` |
| `:many` | `([]Row, error)` |
| `:exec` | `error` (ignores rows affected) |
| `:execrows` | `(int64, error)` — rows affected count |
| `:execresult` | `(pgconn.CommandTag, error)` — full tag for `INSERT 0 1` |
| `:batchexec` | For pgx batch operations |
| `:copyfrom` | For pgx `CopyFrom` |

### pgx Connection Pool

```go
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	// ✅ DO: tune pool size
	config.MaxConns = 25                      // max connections in pool
	config.MinConns = 5                       // minimum idle connections to keep
	config.MaxConnLifetime = 1 * time.Hour    // recycle connections older than this
	config.MaxConnIdleTime = 30 * time.Minute // close idle connections after this

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// ✅ DO: ping to verify connectivity at startup
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
```

### Store Wrapper — Always Wrap Generated Code

```go
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"myapp/internal/store/db"       // sqlc generated
)

type Store struct {
	*db.Queries           // ✅ embed generated code
	pool *pgxpool.Pool    // ✅ keep pool for transaction methods
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: db.New(pool),
		pool:    pool,
	}
}

// ✅ DO: transaction helper — wraps business logic in a TX
func (s *Store) WithTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // safe: Rollback is a no-op if already committed

	q := s.Queries.WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
```

**Trap:** `tx.Rollback(ctx)` after `tx.Commit(ctx)` is a no-op in pgx v5. But calling `tx.Commit(ctx)` after `tx.Rollback(ctx)` returns an error. The pattern above is safe.

### Nullable Columns

```go
// ✅ DO: sqlc with emit_pointers_for_null_types: true generates:
type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Bio       *string    `json:"bio"`       // NULL → nil in Go
	DeletedAt *time.Time `json:"deleted_at"` // NULL → nil in Go
}

// ❌ DO NOT: use sql.NullString manually — sqlc handles this
// ❌ DO NOT: leave nullable columns as non-nullable strings — pgx will error
```

### Migrations (golang-migrate)

```go
import (
	"embed"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(dsn string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("iofs: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// ❌ DO NOT: run migrations in main() with fatal on ErrNoChange — that's a normal state
// ❌ DO NOT: use file-based migration source in Docker — embed is always available
```

---

## 10. Common AI Hallucinations & DeepSeek-Specific Mistakes

### chi APIs That DO NOT Exist

| ❌ Hallucinated API | ✅ Real API / Workaround |
|---------------------|-------------------------|
| `chi.Param(r, "id")` | `chi.URLParam(r, "id")` |
| `chi.RouteContext(ctx)` | `chi.RouteContext(r.Context())` |
| `chi.Middleware` (package) | `"github.com/go-chi/chi/v5/middleware"` (separate sub-package) |
| `chi.Get(r.Context(), key)` | `r.Context().Value(key)` — standard Go |
| `chi.Bind` or `chi.Decode` | Use `json.NewDecoder(r.Body).Decode()` |
| `chi.ParseBody` | Not a thing — use `json.Decode`, `r.ParseForm`, `r.ParseMultipartForm` |
| `chi.Redirect` | `http.Redirect(w, r, url, code)` |
| `chi.Static` or `chi.FileServer` | `http.FileServer(http.Dir("."))` / use `http.ServeFile` |
| `r.Get("/path", handler).Name("name")` | chi has no named routes — use a custom map |
| `r.URLParam("id")` (on the router) | `chi.URLParam(r, "id")` — it's a package function, not a method on `*chi.Mux` |

### chi v5 Specific Traps

```go
// ❌ chi v4 compatibility trap: RouteContext is now inside middleware only
// chi v4: rctx := chi.RouteContext(r.Context())
// chi v5: SAME — but the context structure changed internally.
// Always use: chi.RouteContext(r.Context()) to get the RouteContext.

// ❌ chi.Get(r.Context(), ...) — this is a VEHICLE context method, not chi

// ❌ r.Context().Value(chi.RouteCtxKey) — do NOT use this. Always use chi.RouteContext().

// ✅ DO: use chi.URLParam(r, key) for URL parameters — never dig into context manually
// ✅ DO: use chi.URLParamFromCtx(ctx, key) in middleware that uses r.WithContext() 
//         BEFORE passing to the handler
```

### Handler Traps DeepSeek Frequently Makes

```go
// ❌ TRAP 1: Returning before checking response
func CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := service.CreateUser(r.Context(), input)
	// ❌ DeepSeek often writes the response first, then checks error
	// encodeJSON(w, 201, user)    // WRONG ORDER
	// if err != nil { ... }

	// ✅ CORRECT:
	if err != nil {
		writeError(w, r, err)
		return
	}
	encodeJSON(w, http.StatusCreated, user)
}

// ❌ TRAP 2: Not calling r.Body.Close()
// ❌ TRAP 3: Closing r.Body before json.Decode
// ❌ TRAP 4: Using _ for path params (chi uses {param} syntax)
// r.Route("/api/v1/users/_", ...)  // this matches literal underscore, not a placeholder
// ✅ r.Route("/api/v1/users/{userID}", ...)
```

### Goroutine Leaks in Handlers

```go
// ❌ TRAP: Starting goroutine without lifecycle management
func FireAndForget(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)

	// ❌ This goroutine uses r.Context() which will be cancelled
	//    as soon as the handler returns!
	go func() {
		sendEmail(r.Context(), user) // context already cancelled → send fails silently
	}()

	// ❌ No error handling, no timeout, no lifecycle
}

// ✅ CORRECT: Derive a new context with its own deadline
func FireAndForgetCorrect(w http.ResponseWriter, r *http.Request) {
	// Capture values before the handler exits
	userID := user.ID
	email := user.Email

	w.WriteHeader(http.StatusAccepted)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// ✅ Use values captured from the request, not the request itself
		if err := sendEmail(ctx, userID, email); err != nil {
			slog.Error("background task failed", "userID", userID, "err", err)
		}
	}()
}
```

### Import Traps

```go
// ❌ DeepSeek frequently gets these imports wrong:

// ❌ "github.com/go-chi/chi/middleware"         → v4 path; v5 needs /v5/
// ❌ "github.com/go-chi/chi/v5/middlewares"     → NOT "middlewares"
// ❌ "github.com/go-chi/cors/v5"                → go-chi/cors is versioned separately, no /v5/
// ✅ "github.com/go-chi/chi/v5/middleware"       → correct
// ✅ "github.com/go-chi/cors"                    → correct (check go.mod for actual version)

// ❌ "io/ioutil"                                → deprecated since 1.16
// ✅ "io" + "os"                                → ReadAll, ReadFile, etc. moved

// ❌ "golang.org/x/net/context"                 → pre-Go 1.9
// ✅ "context"                                  → standard library since Go 1.7
```

### sqlc Traps

```go
// ❌ DeepSeek frequently invents sqlc annotations that don't exist:
// -- name: CreateUser :insert   → NO, use :one, :exec, :many, etc.
// -- name: GetUser :single      → NO, use :one
// -- name: List :list           → NO, use :many

// ❌ DeepSeek uses Go's `database/sql` with sqlc
// ✅ sqlc with pgx: set sql_package: "pgx/v5" — generates pgxpool/*.pgxpool, not *sql.DB

// ❌ DeepSeek calls sqlc-generated functions with too many/few args
// ✅ The generated signature matches the SQL parameters exactly in order
```

### Context Deadlines — Silent Failures

```go
// ❌ TRAP: Middleware sets 60s timeout, handler calls external API with 90s timeout
// The 60s middleware Timeout wins. The handler's Deadline will never be reached.
// db calls will context deadline exceeded at 60s with no clear error message.

// ✅ DO: structure timeouts hierarchically
// Outer (middleware.Timeout): 60s  ← total request budget
// Handler:                    50s  ← conservative internal deadline
// External API call:          30s  ← individual operation
// DB query:                    10s  ← individual query
```

### Panic Recovery in Goroutines

```go
// ❌ TRAP: middleware.Recoverer only catches panics in the handler goroutine
// Panics in handler-spawned goroutines will CRASH the server

func Handler(w http.ResponseWriter, r *http.Request) {
	go func() {
		// ❌ No recover — crash!
		processFile(filename)
	}()
}

// ✅ CORRECT: every goroutine must have its own recover
func Handler(w http.ResponseWriter, r *http.Request) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("goroutine panic", "recover", rec)
			}
		}()
		processFile(filename)
	}()
}
```

### Final Checklist: Before Deploying AI-Generated Code

- [ ] Every `r.Body.Close()` is called (ideally via `defer r.Body.Close()` after reading)
- [ ] Every `json.NewDecoder` is preceded or followed by body close
- [ ] Every `WriteHeader` is called at most once per response
- [ ] Every `context.WithCancel`/`WithTimeout`/`WithDeadline` has a matching `defer cancel()`
- [ ] Every goroutine has a `recover()` and uses a context NOT tied to the request
- [ ] Every `http.Handler` function checks the error BEFORE writing the response
- [ ] Panic recovery middleware is AFTER logging middleware in the chain
- [ ] Database pool is setting `MaxConns`, `MinConns`, `MaxConnLifetime`, `MaxConnIdleTime`
- [ ] CORS allows specific origins, never `"*"` with `AllowCredentials: true`
- [ ] URL params use `chi.URLParam(r, "name")` not `chi.Param` or context value direct access
- [ ] Imports use `"github.com/go-chi/chi/v5"` (with `/v5/`)
- [ ] No `io/ioutil` imports — use `io` and `os` instead
- [ ] `signal.NotifyContext` with `os.Interrupt` AND `syscall.SIGTERM`
- [ ] `srv.ListenAndServe()` error is checked against `http.ErrServerClosed` before logging as error
- [ ] All nullable DB columns are represented as pointers in Go structs
