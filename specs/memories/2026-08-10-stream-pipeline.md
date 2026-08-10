# Retro — Stream Status & HLS Pipeline (2026-08-10)

**Source:** [Session Log](./2026-08-10-session-log.md)
**PR:** [#18](https://github.com/eparodi/deepcut-live/pull/18)

## What went well

- WebSocket hub pattern (StreamHub) was straightforward to implement following the existing ChatHub pattern
- Stream key persistence fix was a simple schema change + handler update
- HLS proxy with Next.js rewrites worked once paths were correctly mapped

## What went wrong

- SRS `http_hooks` callbacks never fired — spent hours debugging config formats, env vars, vhost matching. Root cause: SRS vhost matching doesn't inherit hooks from `__defaultVhost__` for dynamic vhosts. Added SRS poller as fallback.
- Docker build cache masked code changes — `--no-cache` is essential after backend modifications
- SRS config file naming (`docker.conf` vs `srs.conf`) caused config to not be loaded

## Rules to Add

### To go-chi skill:
1. Third-party webhook bodies may have extra fields — don't use `DisallowUnknownFields()` on webhook handlers
2. Every WebSocket endpoint must be behind auth middleware
3. WebSocket `AcceptOptions` must use `OriginPatterns`, never `InsecureSkipVerify: true`
4. Nil-guard all injected dependencies that may be nil in tests

### To AGENTS.md or new streaming skill:
5. Store generated secrets (stream keys) alongside their hashes for retrieval
6. Verify actual file locations in containers, not just config values
7. Proxy all paths referenced in HLS playlists, not just the playlist URL
8. SRS management API must be on internal ports only
9. Use `docker compose build --no-cache` after Go code changes

## Skill Updates

Updated go-chi/SKILL.md:
- Added "Third-party webhooks" section (no DisallowUnknownFields)
- Added "WebSocket endpoint checklist" (auth, OriginPatterns, nil guards)
