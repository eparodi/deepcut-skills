# Agent Index

> Canonical roster of every agent skill in the repos.
> Last updated: 2026-08-26. Sources: `*/zed/profiles.json` (v1.5.0),
> `*/.agents/skills/*/SKILL.md`.
>
> **Live catalog:** the generated wiki
> (github.com/eparodi/deepcut-skills/wiki) renders every skill's
> description, category, and section outline; this index keeps the
> per-skill metadata the wiki doesn't generate (tiers, tools, placement).
>
> **Hub:** `skills-test/` holds the canonical copies + the two
> meta-roles. Per-repo copies exist for workflow convenience and DO
> drift — see "Duplication & Drift" below. The learning-porter owns
> keeping them in sync (proposals only).
>
> **Model tiers live ONLY in `zed/profiles.json`** — SKILL.md files
> never declare tiers.

## Role agents (skills that define an agent role)

Tiers: F = Flash default, P = Pro default; "heavy" = escalation tier.

| Role | Scope (one line) | Tier (default/heavy) | Tools | Where |
|---|---|---|---|---|
| `pm` | Requirements, user stories, task breakdown, spec lifecycle | F / P | read/write, no terminal | all repos |
| `architect` | Architecture, API contracts, data models, tech decisions | P / — | read/write, no terminal | all repos |
| `ux-designer` | UI design, components, tokens, accessibility, UX copy | F / P | read/write, no terminal | all repos |
| `backend-engineer` | Go/chi API implementation in `backend/` | F / P | full + terminal | hub + per-repo copy |
| `frontend-engineer` | Next.js App Router implementation in `frontend/` | F / P | full + terminal | hub + per-repo copy |
| `mobile-engineer` | Expo/React Native in `mobile/` | F / P | full + terminal | hub + per-repo copy |
| `ai-engineer` | LLM layer: provider client, malformed-response ladder, retries/breaker, token telemetry, cost budget | F / P | full + terminal + fetch | hub + per-repo copy |
| `db-analyst` | Postgres migrations, store/query layer, data integrity, test-DB strategy | P / — | full + terminal | hub + per-repo copy |
| `financial-analyst` | Trading money: risk rules, sizing, capital, PnL/fees, breaker semantics | P / — | read/write + fetch, no terminal | hub + per-repo copy |
| `bot-engineer` | All Go implementation of the trading bot | F / P | full + terminal | per-repo copy |
| `ui-engineer` | Trading-bot dashboard: Go html/template + folder-per-component CSS/JS | F / P | full + terminal | per-repo copy |
| `reviewer` | PR review: correctness, style, security, standards | F / P | read + terminal, no writes | all repos |
| `qa` | Validates features vs spec, runs suites, QA reports | F / — | read + terminal, no writes | all repos |
| `security-engineer` | Security audits, local pentesting, security reports | P / — | read + terminal, no writes | all repos |
| `orchestrator` | Single-thread 4-role loop (PLANNER/CODER/REVIEWER/DEBUGGER) | F / P | full + terminal | all repos |
| `skill-factory` | Creates super-specific agent skills from recurring evidence (gated) | P / — | read/write + terminal | hub only |
| `learning-porter` | Moves learnings across repos: distill, classify, dedupe, export | P / — | read/write, no terminal | hub only |

## Stack skills (loaded alongside a role, not roles themselves)

| Skill | Covers | Where |
|---|---|---|
| `go-chi` | Go/chi backend standards: layering, errors, concurrency, testing, DB patterns | hub + per-repo copy |
| `nextjs` | Next.js App Router standards: RSC, data fetching, caching, layouts | hub + per-repo copy |
| `expo` | Expo/React Native managed-workflow standards | hub + per-repo copy |
| `go-bot` | Trading-bot Go standards: package layout, slog, ticker/concurrency, BoltDB, exchange/LLM clients, risk guards | per-repo copy |
| `deepcut-platform` | Media-platform specifics: SRS media server, River jobs, HLS/recording paths, env vars | per-repo copy |

## Process skills

| Skill | Covers | Where |
|---|---|---|
| `spec-driven` | 4-phase gated feature lifecycle (Requirements → Design → Tasks → Implementation) | all repos |

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
  in the trading-bot repo) — duplication there is not drift, it is placement.

## Model tier policy

Canonical routing table + escalation ladder:
`skills-test/HOW_WE_WORK.md` → "Model Tier Routing".
Always-on rule: skills-test AGENTS.md §10.20.
Pilot tracker: `skills-test/specs/pilot/agent-cost-pilot.md`.
