# AI-Assisted Development with Zed + DeepSeek

A complete set of agent skills and rules for spec-driven development
across a Go/chi + Next.js + Expo monorepo, designed for the Zed editor
with DeepSeek as the model.

## Why This Exists

Writing code with AI agents is fast. Fixing AI-generated mistakes is slow.
This project front-loads quality by enforcing **spec-first planning** and
**role-based guardrails** so misunderstandings get caught in cheap markdown
edits instead of expensive code rewrites.

It simulates a six-person team — PM, Architect, UX Designer, and three
Engineers — each in its own Zed agent thread, coordinated through a shared
`specs/` directory.

## Quick Start

### 1. Copy into your monorepo

```bash
cp -r skills-test/.agents skills-test/AGENTS.md skills-test/HOW_WE_WORK.md skills-test/specs /path/to/your-monorepo/
```

### 2. Configure Zed profiles

Load `zed/profiles.json` into Zed's agent profile config to get
tool-level guardrails per role (PM/Architect/UX Designer can't run
terminal commands; engineers are scoped to their directories).

### 3. Pick a feature, open threads

Open one Zed agent thread per role. Load the corresponding skill:

| Thread | Skill to load | Command |
|--------|--------------|---------|
| PM | `pm` | `@pm` |
| Architect | `architect` | `@architect` |
| UX Designer | `ux-designer` | `@ux-designer` |
| Backend Eng | `backend-engineer` + `go-chi` | `@backend-engineer @go-chi` |
| Frontend Eng | `frontend-engineer` + `nextjs` | `@frontend-engineer @nextjs` |
| Mobile Eng | `mobile-engineer` + `expo` | `@mobile-engineer @expo` |

### 4. Walk through the phases

Follow `HOW_WE_WORK.md`. The gist:

1. **PM** writes requirements + non-goals → human approves
2. **Architect** + **UX Designer** (parallel) design API + data + UI → human approves
3. **PM** breaks into tasks with role assignments
4. **Engineers** implement one task at a time, verifying against acceptance criteria

### 5. Iterate

After each feature, write a retro in `specs/memories/` tracing any
agent mistakes to the specific missing or weak rule, then tighten it.
The rules get better with every feature.

## Per-Feature Walkthrough

Here's exactly what to paste in each Zed agent thread, per phase.

### Phase 1 — PM writes requirements

Open a PM thread, paste:

```
@pm @spec-driven I need to build [describe your feature in 1-3 sentences].
Start at Phase 1 — draft requirements and non-goals first.
```

The PM writes `specs/<feature-slug>.md`. Review it, request changes if
needed, then say "approved."

### Phase 2 — Architect + UX Designer design (parallel)

Open two threads simultaneously:

**Architect:**
```
@architect @spec-driven specs/<feature-slug>.md has approved requirements.
Produce the Design section — API contract, data model, architecture decisions.
```

**UX Designer:**
```
@ux-designer @spec-driven specs/<feature-slug>.md has approved requirements.
Read the Architect's API contract, then produce the UI Design section —
components, states, tokens, accessibility.
```

Both write to the same spec file. Review both sections, then approve.

### Phase 3 — PM breaks into tasks

Back in the PM thread:

```
@pm @spec-driven specs/<feature-slug>.md is approved through Phase 2.
Break it into tasks with role assignments.
```

### Phase 4 — Engineers implement

**Backend first** (must stabilize the API before frontend/mobile consume it):

```
@backend-engineer @go-chi @spec-driven specs/<feature-slug>.md is ready.
Implement task 1 from the checklist.
```

**Once the API is stable,** Frontend and Mobile run in parallel:

```
@frontend-engineer @nextjs @spec-driven specs/<feature-slug>.md API is live.
Implement task 3.
```

```
@mobile-engineer @expo @spec-driven specs/<feature-slug>.md API is live.
Implement task 5.
```

### Quick Reference Card

| Phase | Who | Load these skills | Prompt |
|-------|-----|-------------------|--------|
| 1 | PM | `@pm @spec-driven` | "Draft requirements for X" |
| 2 | Architect | `@architect @spec-driven` | "Design API + data model for spec X" |
| 2 | UX Designer | `@ux-designer @spec-driven` | "Design UI for spec X" |
| 3 | PM | `@pm @spec-driven` | "Break spec X into tasks" |
| 4 | Backend | `@backend-engineer @go-chi @spec-driven` | "Implement task N from spec X" |
| 4 | Frontend | `@frontend-engineer @nextjs @spec-driven` | "Implement task N from spec X" |
| 4 | Mobile | `@mobile-engineer @expo @spec-driven` | "Implement task N from spec X" |

## What's Included

### Always-On Rules (`AGENTS.md`)

Loaded automatically in every agent thread. Covers:

- **DeepSeek-specific guardrails** — hallucination prevention (cite your source, verify before writing), planning enforcement, over-confidence calibration
- **Codebase pattern matching** — follow existing conventions, don't invent new ones silently
- **Ambiguity handling** — never guess business rules, always ask
- **Output/testing discipline** — build after every change, tests for every state
- **Git hygiene** — conventional commits, branching conventions, never commit unasked

### Spec-Driven Workflow (`spec-driven` skill)

A four-phase, gated process with human checkpoints between phases:

```
Requirements → [approve] → Design → [approve] → Tasks → Implement
```

### Six Role Skills

| Role | Skill | Owns | Can Write Code? |
|------|-------|------|:---:|
| **PM** | `pm` | Requirements, user stories, AC, task breakdown, spec lifecycle | ❌ |
| **Architect** | `architect` | System architecture, API contracts, data models, tech choices | ❌ |
| **UX Designer** | `ux-designer` | Wireframes, design system, component specs, accessibility, UX copy | ❌ |
| **Backend Eng** | `backend-engineer` | `backend/` — Go/chi, database, API implementation | ✅ |
| **Frontend Eng** | `frontend-engineer` | `frontend/` — Next.js App Router, Server Components | ✅ |
| **Mobile Eng** | `mobile-engineer` | `mobile/` — Expo/React Native managed workflow | ✅ |

All six roles are framed as **senior** practitioners — they have guardrails
to flag design flaws, not just blindly implement specs.

Planning roles (PM, Architect, UX Designer) produce design artifacts in
`specs/`. They never touch code. Engineer roles implement against those
artifacts and never silently change the contract.

### Three Stack-Specific Skills

| Skill | Stack | When to load |
|-------|-------|-------------|
| `go-chi` | Go + chi v5 router + stdlib `net/http` | Writing any Go code in `backend/` |
| `nextjs` | Next.js App Router + React Server Components | Writing any TSX in `frontend/` |
| `expo` | Expo (managed) + React Native + Expo Router | Writing any TSX in `mobile/` |

Each covers: project layout, layering rules, error handling, testing
conventions, and — critically — a **complete hallucination reference**
listing every fake API DeepSeek commonly fabricates for that stack,
with the correct alternative.

### Configuration

| File | Purpose |
|------|---------|
| `zed/profiles.json` | Agent profiles with tool-level guardrails and model recommendations |
| `HOW_WE_WORK.md` | Day-to-day workflow: thread setup, handoff protocols, parallel vs sequential, model cost levers |
| `specs/memories/README.md` | Retro template for iterating on skills after each feature |

### Reference Research (`docs/`)

Three ~2,000-line reference documents covering every non-obvious pattern,
trap, and best practice for each stack. Used as source material for the
stack skills but kept for deeper reference.

## Project Structure

```
your-monorepo/
├── AGENTS.md                         # Always-on rules (loaded every thread)
├── HOW_WE_WORK.md                    # Multi-role workflow guide
├── zed/
│   └── profiles.json                 # Zed agent profile config
├── specs/                            # Single source of truth
│   ├── <feature-slug>.md             # Living specs (phases 1-4)
│   └── memories/                     # Retros: trace agent mistakes to rules
├── .agents/
│   └── skills/
│       ├── spec-driven/SKILL.md      # Gated spec-first workflow
│       ├── pm/SKILL.md               # Requirements + task breakdown
│       ├── architect/SKILL.md        # API contracts + data models
│       ├── ux-designer/SKILL.md      # UI design + design system
│       ├── backend-engineer/SKILL.md # Go/chi implementation
│       ├── frontend-engineer/SKILL.md# Next.js implementation
│       ├── mobile-engineer/SKILL.md  # Expo/RN implementation
│       ├── go-chi/SKILL.md           # Go/chi stack standards
│       ├── nextjs/SKILL.md           # Next.js stack standards
│       └── expo/SKILL.md             # Expo/RN stack standards
├── backend/                          # Go code
├── frontend/                         # Next.js code
└── mobile/                           # Expo/RN code
```

## Model Strategy

Use DeepSeek-V4 for ~80% of work. Escalate to Claude Opus or GPT-4o when:

- Designing architecture (Architect role)
- Making hard-to-reverse decisions (data model, auth strategy)
- DeepSeek gets stuck in a hallucination loop
- Cross-cutting refactors or security-sensitive changes

PM, UX Designer, Frontend, and Mobile roles rarely need the stronger model.

## The Iteration Loop

This system is designed to improve with use:

1. Run a feature end-to-end
2. The agent makes a mistake
3. Write a retro: what was the mistake? Which rule should have caught it?
4. Tighten that specific rule
5. Commit both the retro and the updated rule
6. Next feature, that mistake doesn't happen

`specs/memories/` holds the traceability chain: "we added this rule
because of this transcript on this date."

## License

MIT — use, modify, and share freely.
