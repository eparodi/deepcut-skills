# Agent Cost Pilot Tracker

**Spec:** `specs/agent-cost-optimization.md` (US4)
**Period:** 7 days, starting 2026-08-16 00:00 UTC
**Classes under pilot:**
- **A** — Mechanical file ops & test runs (apply spec'd edits, run
  build/test/vet, grep + summarize, commit hygiene)
- **B** — Session-log & retro drafting (`specs/memories/` updates)

Everything not in classes A or B follows the normal routing policy
(Pro for judgment work) and is NOT part of this pilot's numbers.

## Price table (operative from 2026-08-16 16:00 UTC)

Peak hours: 01:00–04:00 and 06:00–10:00 UTC. All other hours off-peak.

| Per 1M tokens | Flash off-peak | Flash peak | Pro off-peak | Pro peak |
|---|---|---|---|---|
| Input (cache hit) | $0.007 | $0.014 | $0.022 | $0.044 |
| Input (cache miss) | $0.22 | $0.44 | $0.66 | $1.32 |
| Output | $0.66 | $1.32 | $1.98 | $3.96 |

Source: api-docs.deepseek.com/quick_start/pricing (fetched 2026-08-15).

## Daily tracker

One row per day per class. Token counts come from the DeepSeek
platform usage page; USD = (hit × hit_rate + miss × miss_rate +
output × output_rate) using the row's peak/off-peak split.
[If the platform does not break down peak/off-peak, log the UTC hour
bands of each run in Notes and compute the split manually.]

| Date (UTC) | Class | Tier | Prompt hit | Prompt miss | Output | Peak share | USD | Successes | Escalations | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| 2026-08-16 | A | Flash | | | | | | | | |
| 2026-08-16 | B | Flash | | | | | | | | |
| 2026-08-17 | A | Flash | | | | | | | | |
| 2026-08-17 | B | Flash | | | | | | | | |
| 2026-08-18 | A | Flash | | | | | | | | |
| 2026-08-18 | B | Flash | | | | | | | | |
| 2026-08-19 | A | Flash | | | | | | | | |
| 2026-08-19 | B | Flash | | | | | | | | |
| 2026-08-20 | A | Flash | | | | | | | | |
| 2026-08-20 | B | Flash | | | | | | | | |
| 2026-08-21 | A | Flash | | | | | | | | |
| 2026-08-21 | B | Flash | | | | | | | | |
| 2026-08-22 | A | Flash | | | | | | | | |
| 2026-08-22 | B | Flash | | | | | | | | |

## Baseline

Previous 7 days (2026-08-09 → 2026-08-15) from the DeepSeek platform
usage page. If historical per-model data is unavailable, the first
pilot week doubles as baseline and expansion waits for a second week's
comparison.

| Metric | Baseline value | Source |
|---|---|---|
| Total tokens (all models) | | platform usage page |
| Total USD (old flat rates) | | computed |
| Cache-hit ratio | | platform usage page (if shown) |
| Escalation count | | manual log |

## Exit criteria (gate to expand Flash to more task classes)

1. Flash success rate ≥ 90% (per class)
2. Escalations < 10% of Flash tasks
3. Total USD during the pilot < baseline USD
4. Peak-share visibility: if peak-share is high, evaluate off-peak
   scheduling of non-urgent work as a lever before expanding

## Result

Filled at pilot end:

- [ ] Criteria met → update the routing table in
      `skills-test/HOW_WE_WORK.md` to add the new task classes
- [ ] Criteria NOT met → revise the routing table first; do not
      expand
