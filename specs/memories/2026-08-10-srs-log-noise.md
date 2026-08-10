# 2026-08-10 Retro: SRS Log Noise Reduction

## Summary

Reduced SRS-related log noise in Docker console. The poller was the #1 noise source (~720 WARN lines/hour in dev), followed by chi's request logger. Two pattern violations were identified and fixed.

---

## Learning 1: Services MUST use injected loggers, not global slog

### Problem

The `StreamService` poller used `slog.Warn/Error/Info` (global package functions) while every other component in the codebase (handlers, hubs, other services) used injected `*slog.Logger` instances. This meant:

- The poller couldn't respect a configured log level
- Log output went through the default logger instead of the structured JSON handler
- Tests couldn't control poller verbosity

### Rule Added

**To: `go-chi` skill → Service layer rules**

```markdown
### DO — Services use injected loggers

Every service MUST accept a `*slog.Logger` via constructor injection. Never use the global `slog.Info/Warn/Error` functions inside a service — they bypass the configured handler and log level.

```go
// ✅ Right: injected logger
type MyService struct {
    logger *slog.Logger
}

func NewMyService(..., logger *slog.Logger) *MyService {
    return &MyService{..., logger: logger}
}

func (s *MyService) doWork() {
    s.logger.Info("work done")
}

// ❌ Wrong: global slog
func (s *MyService) doWork() {
    slog.Info("work done")
}
```

Log helpers MUST be nil-safe so tests can pass `nil`:

```go
func (s *MyService) infoLog(msg string, args ...any) {
    if s.logger != nil {
        s.logger.Info(msg, args...)
    }
}
```

---

## Learning 2: Log level should be configurable via env var

### Problem

Chi's `middleware.Logger` logs every HTTP request at INFO level. In development with frequent SRS callbacks and frontend polling, this generates significant noise with no way to suppress it without code changes.

### Rule Added

**To: `go-chi` skill → main.go / configuration**

```markdown
### DO — Provide LOG_LEVEL env var

The application entry point MUST support a `LOG_LEVEL` environment variable with values: `debug`, `info` (default), `warn`, `error`.

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

---

## Learning 3: Docker services need log rotation

### Problem

Both `backend` and `srs` services wrote unbounded logs to Docker's json-file driver. In long-running development sessions this could fill the disk.

### Rule Added

**To: AGENTS.md → Section 5 (Output & Testing) or a new Docker section**

```markdown
### 5.5 Docker Log Rotation

All services in `docker-compose.yml` MUST include log rotation:
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

---

## Cross-reference

- Session log: `specs/memories/2026-08-10-session-log.md`
- PR: https://github.com/eparodi/deepcut-live/pull/21
