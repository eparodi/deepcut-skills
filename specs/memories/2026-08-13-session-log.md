# Session Log — 2026-08-13

Sync: generic single-thread orchestrator skill from deepcut-live
(PR #27, branch `feat/orchestrator-skill`).

## Corrections & Root Causes

| # | Symptom | Root Cause | Fix |
|---|---------|-----------|-----|
| 1 | deepcut-live orchestrator skill hardcoded project facts (Go 1.25, Next.js 16.3, exact paths/commands) | Written for one repo only | Generic skill now discovers the host repo's stack, layout, entrypoints, and test/build commands itself in Phase 0 and records them in a PLAN.md Context block; stack-skill loading is conditional on what Phase 0 finds |
| 2 | ALWAYS-CHECKS referenced deepcut-live-specific examples (`.nvmrc` content, frontend CI claim) | Copied verbatim from the project-local version | Generalized: CI claims require reading the repo's own workflows; Node guard applies only when a version pin exists; AGENTS.md sections referenced as "if present" |
| 3 | skills-test had no `.gitignore` for orchestrator working docs | Repo never needed one before | Added `.gitignore` with anchored `/PLAN.md`, `/LOOP_LOG.md` |
| 4 | README/HOW_WE_WORK said "six threads" in places where the team table lists nine roles | Older wording predating reviewer/qa/security roles | New orchestrator sections consistently say "nine threads" and match the existing tables |

## Follow-ups / Questions

- [ ] Keep the generic skill in sync with the deepcut-live one: future
      corrections that add ALWAYS-CHECKS or loop rules must be mirrored
      to both repos (same commit pattern as the earlier "docs: sync"
      series).
