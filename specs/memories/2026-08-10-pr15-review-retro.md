# Retro: PR #15 review + security audit — 2026-08-10

## What Happened

PR #15 (add security engineer role) was reviewed by both the Reviewer
and Security Engineer roles. The PR was a meta-PR — adding a new
pipeline role rather than application code. Both gates passed cleanly
with only style and operational notes.

## Findings from Reviews

### Reviewer found:

1. **Session log lacks thematic grouping** — Corrections #1-#6 (security
   audit) and #7-#10 (UX fixes) live in the same flat table. Future
   readers won't know they're from two unrelated workstreams. Consider
   adding a "Context" column.

2. **HOW_WE_WORK footnote grammar** — "cross-cutting refactors, and
   security audits" reads slightly off without a parallel structure.

### Security Engineer found:

3. **Runtime recon curl examples could be copy-pasted against production**
   — The `/api/me/stream-key/regenerate` example with `{"confirm":true}`
   would regenerate a real key if run against production.

4. **No "stop and document" rule** — The non-goals say "don't exploit
   beyond confirming" but don't say "stop when confirmed, don't chain."

## Root Cause

Both reviews exposed a common pattern: **meta-artifacts (skills that
describe how to do work) need their own guardrails for how they'll be
used in practice.** The security skill is thorough on what to check, but
could be clearer on *when to stop checking.*

## Rule Updates Needed

### 1. Security skill: Add production-safety guard
**File:** `.agents/skills/security-engineer/SKILL.md`

Add to the Runtime Reconnaissance section:
```markdown
> ⚠️ **Safety:** All curl examples in this section target `localhost`.
> Never run these commands against a production URL. State-changing
> commands (POST, PATCH, DELETE) should be verified against dev/staging
> only.
```

### 2. Security skill: Add "stop and document" rule
**File:** `.agents/skills/security-engineer/SKILL.md`

Add to Non-Goals or as a new rule:
```markdown
- **Stop on confirmation.** When you confirm a vulnerability exists,
  stop probing deeper and document it. Do NOT chain exploits to
  escalate privilege. The goal is to flag issues, not to prove how
  far an attacker could go.
```

### 3. Session log: Add context column convention
**File:** `specs/memories/README.md` or implicit convention

Session logs that mix unrelated workstreams should group corrections
under sub-headings or add a `Context` column. Example:

```markdown
| # | Context | Event | Root Cause | Fix |
|---|---------|-------|-----------|-----|
| 1 | Security | No security role... | ... | ... |
| 7 | UX Fixes | Navbar useEffect... | ... | ... |
```

## What Went Well

- **Merge conflict resolution was fast** — Only one file conflicted
  (session log). The fix was a straightforward merge of both branches'
  entries into one unified log.
- **Both gates passed on first attempt** — No loop-back cycles needed.
  Meta-PRs (adding roles/docs) are inherently low-risk and tend to
  pass review cleanly.
- **Security audit of a security skill found real issues** — The
  meta-audit (auditing the auditing tool) caught operational risks
  that wouldn't have been visible from a code review alone. This
  validates the decision to run Security in parallel with Reviewer
  rather than relying on the Reviewer to catch security concerns.

## Related

- PR: https://github.com/eparodi/deepcut-live/pull/15
- Session log: `specs/memories/2026-08-09-session-log.md`
- Security-engineer role retro: `specs/memories/2026-08-09-security-engineer-role-retro.md`
