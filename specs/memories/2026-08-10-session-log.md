# Session Log — 2026-08-10

## Corrections Made

| # | Issue | Root Cause | Fix | Missing Rule |
|---|-------|-----------|-----|-------------|
| 1 | Stream key not detected by backend | SRS handler read key from `param` field instead of `stream` field | Added `stream` to JSON struct, primary extraction from `stream` with `param` fallback | Match handler body structs to actual API payload fields |
| 2 | `DisallowUnknownFields()` rejected SRS callback | SRS sends extra fields (ip, vhost, app) not in handler struct | Removed `DisallowUnknownFields()` | Be lenient with third-party webhook bodies that have extra fields |
| 3 | Stream key regenerated every page load | Raw key never stored in DB, frontend auto-generated on every visit | Added `stream_key` column, store raw key alongside hash, return in `/api/me` | Never discard generated secrets; persist them for retrieval |
| 4 | Docker build cache stale after code changes | `go build` step was cached, old binary deployed | Used `--no-cache` flag | Always use `--no-cache` when deploying code changes to Docker |
| 5 | SRS config loaded as `docker.conf` not `srs.conf` | ossrs/srs:5 image reads `docker.conf` by default | Mounted config as `srs.conf` (which SRS also reads via `--config`) | Verify which config file the Docker image actually loads |
| 6 | SRS `http_hooks` not firing for `localhost` vhost | OBS connects as `vhost=localhost` but hooks only on `__defaultVhost__` | Added explicit `vhost localhost` + SRS poller as fallback | SRS vhost matching is not inherited; configure per-vhost |
| 7 | HLS files at wrong path in SRS | SRS default path `./objs/nginx/html/live/` vs configured `/data/hls/` | Used actual SRS default path | Verify actual file locations in the container, not just config |
| 8 | HLS playlist absolute paths break proxy | SRS generates `/live/...` paths in `.m3u8`, proxy only handled `/hls/...` | Added `/live/:path*` proxy rewrite | Proxy all paths referenced in HLS playlists, not just the playlist URL |
| 9 | WebSocket proxy fails through Next.js | Next.js dev rewrites don't proxy WebSocket upgrades | Connected directly to backend on port 8081 | Next.js rewrite proxying does not support WebSocket in dev mode |
| 10 | Chat WebSocket has no authentication | Route registered outside auth middleware | Moved to auth group, read userID from context | Every WebSocket endpoint must be behind auth |
| 11 | `InsecureSkipVerify: true` disables Origin validation | Copied from example without understanding implications | Replaced with explicit `OriginPatterns` | Never use `InsecureSkipVerify` in production; validate Origin headers |
| 12 | SRS `http_api` exposed on public port 8080 | Shared port with `http_server` | Moved `http_api` to internal port 1985 | Management APIs must not be on publicly-exposed ports |
| 13 | Nil hub panic in service tests | `StreamService.hub` accessed without nil check | Added `if s.hub != nil` guards | All service dependencies that may be nil in tests need nil checks |
| 14 | Poller silently discards errors | Copy-paste convenience, no logging | Added `slog.Warn`/`slog.Error` for all error paths | Per go-chi skill: "No bare error discards" |
| 15 | Test compilation failures after signature change | `CreateStream` signature changed but mocks not updated | Updated all mock structs and call sites | Verify call sites before committing signature changes |

## Follow-ups

- [ ] Write migration rollback for `000002_add_stream_key`
- [ ] Add integration test for full OBS → SRS → Backend → Frontend pipeline
- [ ] Evaluate WebRTC for sub-second latency (v2)
