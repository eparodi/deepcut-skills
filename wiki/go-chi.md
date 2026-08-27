# go-chi

> **Generated** from `.agents/skills/go-chi/SKILL.md` — do not edit by hand.
> Source: [SKILL.md](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md)

Go/chi backend development standards — layering, error handling, concurrency, state management, testing, database patterns, and common AI traps. Load when writing Go code in the backend/ directory.

**Category:** stack · **Tags:** go, backend

## Sections

- [Project Layout](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#project-layout)
  - [Layer Rules (apply to any layout)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#layer-rules-apply-to-any-layout)
- [Logging](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#logging)
- [State & Mutable Globals](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#state--mutable-globals)
- [Enums & Magic Strings](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#enums--magic-strings)
- [chi Router Patterns](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#chi-router-patterns)
  - [DO — Standard route registration](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--standard-route-registration)
  - [DO NOT — Fabricate chi APIs](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do-not--fabricate-chi-apis)
  - [DO — Middleware ordering matters](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--middleware-ordering-matters)
- [net/http Patterns](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#nethttp-patterns)
  - [DO — Handler signature](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--handler-signature)
  - [DO — All four server timeouts](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--all-four-server-timeouts)
  - [DO — Exit once, in main (`run()` pattern)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--exit-once-in-main-run-pattern)
- [Goroutine Lifetime](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#goroutine-lifetime)
- [Error Handling](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#error-handling)
  - [DO — Error types](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--error-types)
  - [DO — Match error kinds before fallback behavior](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--match-error-kinds-before-fallback-behavior)
  - [DO — Handle each error once](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--handle-each-error-once)
  - [DO NOT — Error anti-patterns](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do-not--error-anti-patterns)
- [Testing](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#testing)
  - [DO — Table-driven handler tests](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--table-driven-handler-tests)
  - [DO — Integration tests with httptest.NewServer](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--integration-tests-with-httptestnewserver)
  - [DO NOT — Testing mistakes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do-not--testing-mistakes)
- [JSON Encoding/Decoding](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#json-encodingdecoding)
  - [DO — Struct tags reference](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--struct-tags-reference)
  - [DO — Match response shapes to the frontend contract](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--match-response-shapes-to-the-frontend-contract)
  - [DO NOT — JSON mistakes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do-not--json-mistakes)
- [Context Propagation](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#context-propagation)
- [Database (pgx / sqlc)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#database-pgx--sqlc)
  - [DO — pgx error matching and row iteration](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--pgx-error-matching-and-row-iteration)
  - [DO — Guard degenerate lookup keys](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--guard-degenerate-lookup-keys)
  - [DO — Store wrapper with transactions (sqlc)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do--store-wrapper-with-transactions-sqlc)
  - [DO NOT — SQL traps](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do-not--sql-traps)
- [Interfaces](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#interfaces)
- [Common AI Hallucinations — Import Traps](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#common-ai-hallucinations--import-traps)
- [WebSocket (nhooyr.io/websocket)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#websocket-nhooyriowebsocket)
  - [Hub Pattern — Room-based broadcast](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#hub-pattern--room-based-broadcast)
  - [Connection lifecycle — the invariant that prevents leaks](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#connection-lifecycle--the-invariant-that-prevents-leaks)
  - [DO NOT — WebSocket traps](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#do-not--websocket-traps)
  - [Checklist for new WebSocket endpoints](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#checklist-for-new-websocket-endpoints)
- [Third-Party Webhooks](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#third-party-webhooks)
- [Subprocess Hygiene (ffmpeg and friends)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#subprocess-hygiene-ffmpeg-and-friends)
- [Nil Guards for Injected Dependencies](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#nil-guards-for-injected-dependencies)
- [Pre-Deploy Checklist](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-chi/SKILL.md#pre-deploy-checklist)

