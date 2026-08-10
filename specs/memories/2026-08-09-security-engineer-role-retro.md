# Retro: security-engineer-role — 2026-08-09

## What Happened

During a security audit of the DeepCut Live application, we discovered
the multi-role AI pipeline had no dedicated security role. The QA role
had a cursory "Security quick-check" section (checking for `.env.example`
entries and `fmt.Sprintf` in SQL), but no comprehensive security auditing.

Meanwhile, the application had **12 security vulnerabilities** including
2 critical (WebSocket chat impersonation, WebSocket origin verification
disabled), 3 high (no rate limiting, ephemeral JWT keys, predictable
SRS secret), 4 medium, and 3 low.

## Root Cause

The original pipeline design had 6 roles (PM, Architect, UX Designer,
Backend Eng, Frontend Eng, Mobile Eng) plus Reviewer and QA for PR
verification. But no role was responsible for security-specific concerns
like:

- OWASP Top 10 / API Security Top 10 compliance
- Penetration testing on local deployments
- Cryptographic standards enforcement
- WebSocket security
- Rate limiting review
- Security headers verification
- Dependency vulnerability scanning

The Reviewer role touches on security (`🔴 Critical` includes "Security
vulnerability") but the Reviewer is a generalist — they don't have the
specialized knowledge base to catch everything a dedicated security
auditor would.

## What We Did

1. **Created `.agents/skills/security-engineer/SKILL.md`** — a comprehensive
   security role with:
   - Full OWASP Top 10 and API Security Top 10 knowledge base
   - Common attack vectors with detection patterns
   - Go and TypeScript/Next.js-specific security patterns
   - Static analysis and runtime reconnaissance procedures
   - Structured report format with severity levels

2. **Updated the spec-driven workflow** — Changed Phase 5 from "QA only"
   to "QA + Security (parallel)". Both roles run simultaneously; both
   must pass before the PR can proceed.

3. **Updated HOW_WE_WORK.md** — Added Security Engineer to the role table
   and parallel execution model.

## Rule Added

**New skill file:** `.agents/skills/security-engineer/SKILL.md`

Key design decisions:
- Runs in parallel with QA (no added pipeline latency)
- Has built-in security knowledge (OWASP, API security, attack patterns)
- Can perform active penetration testing on local deployments
- Produces actionable, severity-ranked reports
- Never writes implementation code (non-goal)

**Updated workflow:** `.agents/skills/spec-driven/SKILL.md` Phase 5

Changed from sequential QA-only gate to parallel QA + Security gate.
Both `[QA_PASS]` and `[SECURITY_PASS]` required before final approval.

## Learnings for Future Roles

1. **Every pipeline gate should be a dedicated role.** Bolting security
   checks onto QA or Reviewer roles leads to surface-level checks.
   Dedicated roles carry their own knowledge base and procedures.

2. **Knowledge-dense roles need built-in reference material.** The
   security engineer skill includes OWASP categories, attack vectors,
   cryptographic standards, and language-specific patterns. This
   prevents the agent from relying on general knowledge (which may be
   outdated or incomplete for newer models).

3. **Parallel execution is free in CI time.** QA and Security both
   inspect the same codebase; running them sequentially adds latency
   with no benefit. The parallel gate pattern (both must pass) is the
   right design.

4. **Security audits should be routine, not exceptional.** By making the
   security engineer a standard pipeline role, every PR gets audited —
   not just the ones someone remembers to check.

## Related

- Session log: `specs/memories/2026-08-09-session-log.md`
- PR: https://github.com/eparodi/deepcut-live/pull/15
