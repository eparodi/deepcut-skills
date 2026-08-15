# Agent Index — noir-hq

> Canonical roster of every agent skill in the noir-hq repos.
> Last updated: 2026-08-15. Sources: `*/zed/profiles.json` (v1.4.0),
> `*/.agents/skills/*/SKILL.md`.
>
> **Hub:** `skills-test/` holds the canonical copies + the two
> meta-roles. Per-repo copies exist for workflow convenience and DO
> drift — see "Duplication & Drift" below. The learning-porter owns
> keeping them in sync (proposals only).

## Role agents (skills that define an agent role)

Tiers: F = Flash default, P = Pro default; "heavy" = escalation tier.

| Role | Scope (one line) | Tier (default/heavy) | Tools | Where |
|---|---|---|---|---|
| `pm` | Requirements, user stories, task breakdown, spec lifecycle | F / P | read/write, no terminal | all 3 repos |
| `architect` | Architecture, API contracts, data models, tech decisions | P / — | read/write, no terminal | all 3 repos |
| `ux-designer` | UI design, components, tokens, accessibility, UX copy | F / P | read/write, no terminal | all 3 repos |
| `backend-engineer` | Go/chi API implementation in `backend/` | F / P | full + terminal | skills-test, deepcut-live |
| `frontend-engineer` | Next.js App Router implementation in `frontend/` | F / P | full + terminal | skills-test, deepcut-live |
| `mobile-engineer` | Expo/React Native in `mobile/` | F / P | full + terminal | skills-test, deepcut-live |
| `ai-engineer` | LLM layer: provider client, malformed-response ladder, retries/breaker, token telemetry, cost budget | F / P | full + terminal + fetch | skills-test, deepcut-binance-bot |
| `financial-analyst` | Trading money: risk rules, sizing, capital, PnL/fees, breaker semantics | P / — | read/write + fetch, no terminal | skills-test, deepcut-binance-bot |
| `bot-engineer` | All Go implementation in the bot repo | F / P | full + terminal | deepcut-binance-bot |
| `ui-engineer` | Bot dashboard: Go html/template + folder-per-component CSS/JS | F / P | full + terminal | deepcut-binance-bot |
| `reviewer` | PR review: correctness, style, security, standards | F / P | read + terminal, no writes | all 3 repos |
| `qa` | Validates features vs spec, runs suites, QA reports | F / — | read + terminal, no writes | all 3 repos |
| `security-engineer` | Security audits, local pentesting, security reports | P / — | read + terminal, no writes | all 3 repos |
| `orchestrator` | Single-thread 4-role loop (PLANNER/CODER/REVIEWER/DEBUGGER) | F / P | full + terminal | all 3 repos |
| `skill-factory` | Creates super-specific agent skills from recurring evidence (gated) | P / — | read/write + terminal | skills-test |
| `learning-porter` | Moves learnings across repos: distill, classify, dedupe, export | P / — | read/write, no terminal | skills-test |

## Stack skills (loaded alongside a role, not roles themselves)

| Skill | Covers | Where |
|---|---|---|
| `go-chi` | Go/chi backend standards: layering, errors, concurrency, testing, DB patterns | skills-test, deepcut-live |
| `nextjs` | Next.js App Router standards: RSC, data fetching, caching, layouts | skills-test, deepcut-live |
| `expo` | Expo/React Native managed-workflow standards | skills-test, deepcut-live |
| `go-bot` | Bot repo Go standards: package layout, slog, ticker/concurrency, BoltDB, Binance/DeepSeek clients, risk guards | deepcut-binance-bot |
| `deepcut-platform` | deepcut-live platform specifics: SRS media server, River jobs, HLS/recording paths, env vars | deepcut-live |

## Process skills

| Skill | Covers | Where |
|---|---|---|
| `spec-driven` | 4-phase gated feature lifecycle (Requirements → Design → Tasks → Implementation) | all 3 repos |

## Duplication & Drift

- 10 skills exist in 2+ repos as separate files: `pm`, `architect`,
  `ux-designer`, `reviewer`, `qa`, `security-engineer`,
  `orchestrator`, `spec-driven` (3 repos each); `ai-engineer`,
  `financial-analyst` (2 repos each).
- The per-repo registry files (`zed/profiles.json`) and
  `HOW_WE_WORK.md` had already diverged before 2026-08-15 (different
  role sets, versions 1.3.0 / 1.2.0 / 1.0.0) — the skill files are
  subject to the same drift risk. The learning-porter is the
  designated syncer; it proposes, never auto-applies.
- Stack skills are deliberately repo-scoped (`go-bot` only makes sense
  in the bot repo) — duplication there is not drift, it is placement.

## Model tier policy

Canonical routing table + escalation ladder:
`skills-test/HOW_WE_WORK.md` → "Model Tier Routing".
Always-on rule: skills-test AGENTS.md §10.20.
Pilot tracker: `skills-test/specs/pilot/agent-cost-pilot.md`.
