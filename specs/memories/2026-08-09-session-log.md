# Session Log — 2026-08-09

## Session Summary
Created the Security Engineer role and performed initial security audit of DeepCut Live.

## Corrections & Root Causes

| # | Event | Root Cause | Fix |
|---|-------|-----------|-----|
| 1 | No security role existed in the workflow | Security auditing was a gap in the pipeline | Created `security-engineer` skill with comprehensive knowledge base |
| 2 | WebSocket chat has no auth | User identity from untrusted query params | Flagged as VULN-01 (Critical) — needs JWT auth on WS upgrade |
| 3 | WebSocket origin verification disabled | `InsecureSkipVerify: true` | Flagged as VULN-02 (Critical) — needs origin checking |
| 4 | No rate limiting anywhere | Missing middleware | Flagged as VULN-03 (High) |
| 5 | Ephemeral JWT keys on restart | Auto-generated keys with no persistence | Flagged as VULN-04 (High) |
| 6 | Predictable SRS secret default | `dev-srs-secret` hardcoded fallback | Flagged as VULN-05 (High) |

## Questions / Follow-Ups

- [ ] Should rate limiting be implemented via chi middleware or a reverse proxy?
- [ ] Should the security audit run on every PR or only on release branches?
- [ ] Do we need a dependency vulnerability scanner in CI (e.g., nancy, govulncheck)?
- [ ] Should we add a `.env.example` file with documented security requirements?

## Files Changed

- **Created:** `.agents/skills/security-engineer/SKILL.md` — New security engineer skill
- **Modified:** `.agents/skills/spec-driven/SKILL.md` — Updated Phase 5 to run QA + Security in parallel
- **Modified:** `HOW_WE_WORK.md` — Added Security Engineer to role table and parallel execution docs
