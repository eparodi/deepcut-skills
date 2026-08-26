# Skill Catalog

> **Generated** from `.agents/skills/*/SKILL.md` + `wiki/catalog.json` — do not edit by hand.
> Regenerate with `go run ./tools/wiki-gen`, verify with `go test ./...`.

The agent skills of the deepcut-skills hub, cataloged from their SKILL.md files.

## Role agents (15)

| Skill | Description | Page |
|---|---|---|
| `ai-engineer` | AI Engineer — implements the LLM layer of a repo: provider client (OpenAI-compatible APIs, DeepSeek as the reference platform), malformed-response ladder (repair/re-ask then HOLD), retries/backoff/circuit breaker, prompt construction, token telemetry, and the cost-budget mechanism. Implementation-only role: works from approved specs, test-first. Never changes API contracts or risk rules unilaterally. | [ai-engineer](ai-engineer) |
| `architect` | Architect — owns system architecture, API contracts, data models, and technology decisions. Never writes implementation code. Produces design artifacts that engineers build against. | [architect](architect) |
| `backend-engineer` | Backend Engineer — owns Go/chi implementation in the monorepo backend/. Implements API contracts from specs. Must publish stable endpoints before Frontend/Mobile consume them. Never changes API contracts unilaterally. | [backend-engineer](backend-engineer) |
| `db-analyst` | DB Analyst — owns Postgres migration mechanics, the store/query layer, data integrity, and test-DB strategy for the backend. Reviews the Architect's data model for mechanical safety, never designs tables. | [db-analyst](db-analyst) |
| `financial-analyst` | Financial Analyst — owns the money behavior of trading systems: risk rules, position sizing, capital management, PnL/fee accounting, breaker semantics, and trading soundness. Reviews specs and live trading data for financial soundness; enforces serious-source discipline (academic/institutional only, no guru/scam content). Never writes implementation code. Load for any spec/design/risk decision that touches the trading money. | [financial-analyst](financial-analyst) |
| `frontend-engineer` | Frontend Engineer — owns Next.js App Router implementation in the monorepo frontend/. Builds UI against stable API contracts from the Backend Engineer. Never changes API contracts or touches backend/ or mobile/. | [frontend-engineer](frontend-engineer) |
| `learning-porter` | Learning Porter — distills session corrections into rules, classifies generic vs repo-specific, dedupes, and ports learnings across the repos and beyond. Proposes, never auto-applies. | [learning-porter](learning-porter) |
| `mobile-engineer` | Mobile Engineer — owns Expo/React Native (managed workflow) implementation in the monorepo mobile/. Builds against stable API contracts from the Backend Engineer. Never changes API contracts or touches backend/ or frontend/. | [mobile-engineer](mobile-engineer) |
| `orchestrator` | Single-thread Master Orchestrator — simulates a 4-role team (PLANNER/CODER/REVIEWER/DEBUGGER) in one agent thread and loops until PLAN.md is fully checked off. Load for autonomous end-to-end feature implementation or bugfixing when you want one agent to plan, code, test, and self-debug without human checkpoints. | [orchestrator](orchestrator) |
| `pm` | Project Manager — owns requirements, user stories, acceptance criteria, task breakdown, and spec lifecycle. Never writes code. Routes all ambiguity back to artifact owners. | [pm](pm) |
| `qa` | QA Engineer — validates completed features against their specification. Runs test suites, verifies API contracts, inspects UI states, and produces structured QA reports. Never writes implementation code. | [qa](qa) |
| `reviewer` | Code Reviewer — reviews PRs for correctness, style, security, and adherence to project standards. Produces structured review reports with severity levels. Never writes implementation code. | [reviewer](reviewer) |
| `security-engineer` | Security Engineer — audits the application for security vulnerabilities, performs penetration testing on local deployments, and enforces security standards. Runs in parallel with QA. Produces structured security reports. Never writes implementation code. | [security-engineer](security-engineer) |
| `skill-factory` | Skill Factory — creates super-specific agent skills from recurring evidence, gated, deduped, YAML-validated, and human-approved. Never speculative. | [skill-factory](skill-factory) |
| `ux-designer` | UX Designer — research-driven owner of UI design: screens, design system, spacing, responsive behavior, accessibility, and UX copy. Grounds every decision in documented principles from major design systems (NN/g, Apple HIG, Material, WCAG, Laws of UX). Coordinates with the Architect on the API/UI boundary. Never writes implementation code. | [ux-designer](ux-designer) |

## Stack skills (3)

| Skill | Description | Page |
|---|---|---|
| `expo` | Expo/React Native (managed workflow) development standards — project layout, Expo Router, platform-specific patterns, testing, and managed workflow boundaries. Load when writing React Native code in the mobile/ directory. | [expo](expo) |
| `go-chi` | Go/chi backend development standards — layering, error handling, concurrency, state management, testing, database patterns, and common AI traps. Load when writing Go code in the backend/ directory. | [go-chi](go-chi) |
| `nextjs` | Next.js App Router development standards — Server Components, data fetching, route handlers, caching, layout patterns, and platform-specific AI traps. Load when writing TypeScript/React code in the frontend/ directory. | [nextjs](nextjs) |

## Process skills (1)

| Skill | Description | Page |
|---|---|---|
| `spec-driven` | Design and implement features using a specification-first approach with gated phases. Draft requirements, get approval, design architecture, get approval, break down tasks, then implement one task at a time with verification against acceptance criteria. | [spec-driven](spec-driven) |

## Regenerating & publishing

- Regenerate: `go run ./tools/wiki-gen` (then `go test ./...`)
- Publish: `./tools/wiki-gen/publish.sh`
