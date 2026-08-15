# Agent Specialization Review — findings & proposals

**Status:** Partially implemented (groups A + B done 2026-08-15; C/D/E pending user decisions)
**Created:** 2026-08-15
**Updated:** 2026-08-15 — tier-directive + db-analyst recorded; A/B marked done
**Method:** 3 parallel read-only reviews (all 46 SKILL.md files read;
byte-level diffs of the 16 duplicated skills; grep for tier
declarations). Roster: `skills-test/AGENT_INDEX.md`.

## Resolved decisions (2026-08-15)

1. **Model tiers live ONLY in `zed/profiles.json`** — SKILL.md files
   never declare tiers. The skill-factory template now pairs each new
   skill with a profile entry instead of a tier line.
2. **`db-analyst` role adopted** (user request, passed the factory
   gates): owns Postgres migrations, store/query layer, data
   integrity, test-DB strategy; Pro default; skill + profile entries
   in skills-test and deepcut-live. Test task produced real findings:
   migrations 000001/000002 lack `IF NOT EXISTS` idempotency, no down
   migrations exist, all 4 postgres adapters correctly check
   `rows.Err()`.

## Universal finding

~~Zero skills declare a `Model tier:` line.~~ **Resolved:** tiers are a
profile property — see A1 below. Registry gaps were filled so every
role in each repo now has a profile entry with tiers (v1.5.0).

## Action groups (proposals for your selection)

### A. Universal hygiene — mechanical, low risk [DONE 2026-08-15]

| # | Action | Files | Why |
|---|---|---|---|
| A1 | Tiers defined in profiles (user directive): every role now has a profile entry in its repo's registry with Flash/Pro tiers — added missing profiles (skills-test: reviewer/qa/security; deepcut-live: ai-engineer/financial-analyst/reviewer/qa/security; bot: ui-engineer); registries v1.5.0 | all 3 `zed/profiles.json` | Routing enforceable per role; skills stay tier-free |
| A2 | Stale content fixed: bot architect risk ownership → financial-analyst; bot reviewer examples; go-bot project layout (14 real packages); live expo/nextjs nvm + nonexistent `mobile/` refs | bot + live repos | Stale facts poison threads; flagged by all 3 reviewers |

### B. Dedupes — reduces context cost per thread [DONE 2026-08-15]

| # | Action | Files | Why |
|---|---|---|---|
| B1 | Deleted the `## Operating Inside the Orchestrator` block from backend/frontend/mobile/ai-engineer + bot ai-engineer/bot-engineer (×9) — permission override + completion markers now live once in the compressed spec-driven pipeline | 9 files across 3 repos | Triplication paid for on every thread load |
| B2 | Collapsed the two pipeline definitions: spec-driven's Post-Approval section compressed from ~190 lines (incl. 80-line bash script) to ~45; orchestrator remains the single-thread alternative; one pipeline definition | all 3 `spec-driven` copies | Two competing definitions of the same loop |
| B3 | qa's `Security quick-check` → "Security flags": QA consumes the security-engineer report instead of re-running greps | hub + live + bot `qa` | Same checks maintained twice |
| B4 | security-engineer's generic OWASP/crypto knowledge base → `skills-test/docs/references/security-knowledge-base.md`, loaded on demand | hub + live `security-engineer` | Generic knowledge paid for on every load |

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
