# Agent Specialization Review — findings & proposals

**Status:** Draft — awaiting user selection
**Created:** 2026-08-15
**Method:** 3 parallel read-only reviews (all 46 SKILL.md files read;
byte-level diffs of the 16 duplicated skills; grep for tier
declarations). Roster: `skills-test/AGENT_INDEX.md`.

## Universal finding

**Zero skills declare a `Model tier:` line.** The tiered routing policy
(§10.20) and the skill-factory template require it; without it, the
Flash/Pro routing is unenforceable per skill. This is the cheapest,
highest-value fix: one line per role skill.

## Action groups (proposals for your selection)

### A. Universal hygiene — mechanical, low risk [propose: do now]

| # | Action | Files | Why |
|---|---|---|---|
| A1 | Add `Model tier:` line to every role skill (16 roles; stack skills get "N/A — loaded alongside a role") | all repos | Enables §10.20 routing; factory template compliance |
| A2 | Fix stale content: architect's risk ownership (bot: moved to financial-analyst 2026-08-13), reviewer's `us4-chat`/`us2-streams` examples (bot copy), go-bot's project layout (missing `auth`, `backtest`, `bundle`, `dashboard`, `notify`), deepcut-live's `.nvmrc`/`mobile/` refs (no `mobile/` dir exists), spec-driven's HTTP/SQL examples (bot copy uses BoltDB, no HTTP API) | bot + live repos | Stale facts poison threads; flagged by all 3 reviewers |

### B. Dedupes — reduces context cost per thread [propose: do now]

| # | Action | Files | Why |
|---|---|---|---|
| B1 | Delete the verbatim `## Operating Inside the Orchestrator` block from backend/frontend/mobile-engineer (×3, ~40 lines each) — keep once in orchestrator | deepcut-live + skills-test | Pure triplication; tokens on every thread load |
| B2 | Collapse the two pipeline definitions: spec-driven's `## Post-Approval Orchestration` (~190 ln incl. 80-line bash script) duplicates orchestrator's loop — make orchestrator the canonical pipeline; spec-driven keeps phases 0–4 + points at it | all 3 repos | Two competing definitions of the same loop, same completion markers |
| B3 | qa's `Security quick-check` section → replace with "read security-engineer's report"; keep the greps in security-engineer only | bot + hub | Same checks maintained twice |
| B4 | security-engineer's generic OWASP/crypto knowledge-base tables (~70–135 ln) → move to an on-demand reference doc, skill references it | all repos | Generic knowledge every model already has; paid for on every load |

### C. Splits — more specific agents [propose: pick & spec]

Each split must pass the skill-factory evidence gate (recurring task
pattern in session logs/retros). Candidates, strongest first:

| # | Split | From | Rationale / evidence |
|---|---|---|---|
| C1 | `bot-engineer` → `trading-loop-engineer` (ensemble/risk/execute/memory/sampler) + `integrations-engineer` (market + llm clients) + `bot-ops` (config, systemd, deploy-pi.sh, cross-compile) | bot-engineer (174 ln, "all Go code") | "All Go code" is unmanageable; the repo now has auth/backtest/bot/bundle/dashboard/notify packages; deploy-ordering corrections recur in session logs (§10.10) |
| C2 | `media-pipeline` (ffmpeg/SRS/HLS/recording) from go-chi's `Subprocess Hygiene` + deepcut-platform's SRS content | go-chi (888 ln) + deepcut-platform | Media failures recurred repeatedly (skills-test §10.7, §10.8: SRS payload drift, subprocess hygiene); the knowledge is split across two skills today |
| C3 | `go-websocket` from go-chi's 100-line WebSocket section | go-chi | Self-contained topic, only needed by streaming work |
| C4 | `web-ui-patterns` from nextjs's 220-line Component Patterns section | nextjs (854 ln) | Shared by frontend + dashboard UI work; nextjs itself stays routing/data-focused |
| C5 | `eas-build-deploy` from expo's Build & Deploy section | expo (533 ln) | Deploy-only knowledge, loaded only when shipping |

### D. Merges — canonical copies [propose: pick]

| # | Action | Details |
|---|---|---|
| D1 | `ux-designer` 3-way merge | Hub has the PRINCIPLES core; bot copy adds money-semantics exclusions + financial-UX hard rules + dashboard project context; live copy is stale. Merge = hub generic core + per-repo Project Context sections |
| D2 | Decide the `orchestrator` divergence | Live copy replaced Phase 0 with hardcoded `CONTEXT INJECTION` facts (Go/Next/Node versions, port, commands) — unique value but a staleness hazard the file admits; bot copy is hub + trivial tweaks (→ canonical-ize bot). Options: (a) keep CONTEXT INJECTION but make it generated/dated, (b) revert to discovery-based Phase 0 |
| D3 | Canonicalize the 6 byte-identical hub↔live stack/engineer copies (backend/frontend/mobile-engineer, expo, go-chi, nextjs) | They're identical today — keep one source of truth (hub) + per-repo symlink/copy discipline via the learning-porter |

### E. Open decisions

| # | Question |
|---|---|
| E1 | `mobile-engineer` + `expo` exist in deepcut-live but `mobile/` does NOT exist there. Remove from deepcut-live, or is mobile/ planned? |
| E2 | bot `pm`/`spec-driven` never route to `ux-designer` even though the bot repo has one now — update the bot's spec-driven Phase 2 to include UX? |
| E3 | Splits C1–C5: which ones do you want spec'd? (Each becomes a small spec + skill-factory run; evidence for C1/C2 is already in session logs/§10, C3–C5 are structural only.) |

## Recommended order

1. **A1 + A2 + B1–B4** — one pass, Flash-tier mechanical work, ~2h,
   immediate context-cost savings on every future thread.
2. **D1/D2/D3** — hub canonicalization, then porter keeps it synced.
3. **C-splits** — only after the pilot data shows where Flash
   struggles (the splits themselves should not be speculative).
