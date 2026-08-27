# Visual QA: Browser Automation + DeepSeek Vision

**Status:** Approved (requirements gate 2026-08-26; design gate 2026-08-26)
**Owner:** Eliseo
**Created:** 2026-08-26

## Context

DeepSeek released a vision model (`deepseek-v4-flash-vision-exp`) that
accepts images alongside text in the standard OpenAI-compatible Chat
Completions format (verified against api-docs.deepseek.com, 2026-08-26):

- Images ride in a `user` message as a base64 `data:` URL (images in
  system/assistant messages → 400).
- Formats: JPEG, PNG, GIF, WebP — detected from file content, not name.
- Request body ≤ 48 MiB; single base64 image ≤ 32 MiB.
- Images are auto-resized to ~800×800 before inference (max **384
  tokens per image**) — a full-page mobile screenshot costs ≤ 384
  vision tokens.
- Optional `detail` field: `low` downscales to 512×512 (faster/cheaper
  when fine detail doesn't matter); default (`auto`) keeps ~800×800 —
  the right trade for reading UI text.

The hub repo already carries a hard-won lesson (skills-test AGENTS.md
§10.45): *text-only* models can't consume screenshots — visual QA had
to render to a file and pass plain-text references to it. The vision
model removes that limitation for QA screenshots: the agent can now
send the actual pixels.

**Goal:** a cheap visual-QA capability in the skills set. A Go CLI in
`tools/visualqa/` drives a headless Chrome session (via CDP), captures
screenshots at scripted points in a flow, sends them to the vision
model with a QA checklist, and produces a structured report — with
device emulation for desktop, tablet, and mobile (mobile first, as the
most important design target).

## Resolved Decisions (requirements gate, 2026-08-26)

1. **Interaction model — v1 is scripted flows + per-step vision
   verification.** The vision-driven autonomous loop (model picks the
   next action from the screenshot) is an explicit follow-up (VQ-1).
2. **Browser automation dependency — approved.** `github.com/go-rod/rod`
   (CDP client with built-in device presets). The repo is currently
   stdlib-only; this is the one new dependency. Needs network for
   `go get` and a local Chrome/Chromium at run time.
3. **Budget defaults — approved.** `--max-steps 15`, `--max-screenshots
   12`, per-run timeout 5m, retry budget 3.
4. **API key — repo root.** `DEEPSEEK_API_KEY` read from a gitignored
   `.env.visualqa` in the repo root; the `DEEPSEEK_API_KEY` environment
   variable takes precedence. The key goes ONLY in the Authorization
   header — never in prompts, logs, or reports.
5. **Repo layout — single module for v1.** One root `go.mod` hosts both
   `tools/wiki-gen` and `tools/visualqa`; `go test ./...` covers both.
   Splitting `tools/visualqa` into its own module is a follow-up (VQ-2).
6. **Model is user-configurable.** Resolution order: `--model` flag >
   `DEEPSEEK_VISION_MODEL` (process env, then `.env.visualqa`) >
   default `deepseek-v4-flash-vision-exp`.

## Requirements

### User Story 1: Screenshot-and-verify one page (mobile first)

As a QA engineer, I want to open a URL in a headless browser at a
chosen device form factor and get a vision-verified visual QA report,
so I can catch rendering and layout issues that unit tests can't.

**Acceptance Criteria:**
- Given a URL and a device profile (`mobile` | `tablet` | `desktop`,
  default `mobile`), When I run
  `go run ./tools/visualqa --url <url> --device mobile`, Then a
  headless Chrome session launches with that device's emulation
  (viewport, user agent, touch, device scale factor), navigates to the
  URL, waits for load, and captures a viewport screenshot (the pixels
  a user actually sees) saved under `artifacts/visualqa/<run>/`.
- Given the captured screenshot, When it is sent to
  `deepseek-v4-flash-vision-exp` with the QA checklist prompt, Then
  the response parses into structured findings (per-check verdict
  PASS / FAIL / UNCERTAIN with a reason) and a human-readable
  `report.md` is written next to the screenshots.
- Given a report, When I open it, Then every FAIL/UNCERTAIN finding
  references its screenshot and the reason, and the run id, device,
  and URL are visible in the header.

### User Story 2: Multi-step flow with per-step verification

As a QA engineer, I want to script a flow (goto → click → type →
scroll → wait → screenshot) and verify each step visually, so I can QA
interactions, not just a landing page.

**Acceptance Criteria:**
- Given a flow file (`--flow flows/<feature>.json`), When the tool
  runs, Then it executes the steps in order and captures a screenshot
  at every step marked `"capture": true` and at every explicit
  `screenshot` step.
- Given a step that fails to locate its target element, When the tool
  runs, Then it records the step as FAILED with the CDP error, aborts
  the remaining steps (fail-fast), and includes the failure in the
  report.
- Given a flow file with an unknown action or malformed JSON, When the
  tool runs, Then it exits non-zero with a validation error naming the
  step, **without** launching a browser.
- Given console errors or failed network requests during a run, When
  the run completes, Then they appear in the report under a
  Diagnostics section (captured from CDP events — this is the
  "interacting with the browser using events" part).

### User Story 3: Desktop / tablet / mobile presets

As a QA engineer, I want the three form factors available from one
flag so the same flow runs across device classes.

**Acceptance Criteria:**
- Given `--device desktop|tablet|mobile`, When the tool runs, Then the
  browser emulates the corresponding profile (mobile: phone viewport +
  touch + UA; tablet: tablet viewport + touch; desktop: desktop
  viewport, no touch) and the report records the profile used.
- Given a flow file, When I run it once per device, Then each run
  produces its own run directory and report under
  `artifacts/visualqa/`.

### User Story 4: Reliability & cost guardrails

As an operator, I want bounded runs, so a runaway flow or a flaky
model response can't burn tokens or hang.

**Acceptance Criteria:**
- Given a run, When it exceeds `--max-steps` (default 15),
  `--max-screenshots` (default 12), or the per-run timeout (default
  5m), Then it stops immediately, marks the run FAILED with the
  reason, and writes whatever report exists.
- Given a provider error classified retryable (429 / 500 / 503 /
  network timeout), When a vision call fails, Then the tool retries
  with exponential backoff + full jitter up to the retry budget
  (default 3), honoring `Retry-After` when present.
- Given a provider error classified non-retryable
  (400 / 401 / 402 / 422), When a vision call fails, Then the tool
  does NOT retry — it fails the run with the provider's message
  (never re-sends a malformed payload).
- Given an empty or malformed vision response (JSON-mode quirk), When
  it occurs, Then the tool attempts one local repair (strip markdown
  fences, extract the first balanced JSON object) and then one
  re-ask, and if still invalid marks the step UNCERTAIN and continues
  — a single flaky answer never hard-crashes the run.
- Given a run, When it completes, Then the report records the
  screenshots taken and vision tokens used per step, so cost is
  visible.

### User Story 5: Skill integration

As an agent in this repo, I want a `visual-qa` skill that documents
how to run the tool, choose devices, and interpret findings, so any
QA session can use it consistently.

**Acceptance Criteria:**
- Given the skill roster, When the skill lands, Then
  `.agents/skills/visual-qa/SKILL.md` exists, YAML-parses, is listed
  in `AGENT_INDEX.md`, has a profile entry in `zed/profiles.json`,
  and the wiki is regenerated so `go test ./...` passes.
- Given the skill, When a QA session loads it, Then it explains: how
  to install/run the tool, the device presets, the flow-file format,
  the API-key convention, and how to read the report.

### User Story 6: Feature & case authoring in the flow JSON

As a QA engineer, I want a documented flow-file schema that holds one
feature with MULTIPLE named cases, validated strictly, so adding a new
feature or a new case is a consistent, rule-enforced JSON edit.

**Acceptance Criteria:**
- Given the flow schema, When a file declares `feature`, `base_url`,
  and `cases` (1..N, each with a `name` and ordered `steps`), Then it
  loads and runs each case in order — or exactly one case when
  `--case <name>` is given.
- Given a flow file with an unknown action, an unknown field, a
  missing name, or empty steps, When the tool runs, Then it exits
  non-zero naming the offending case/step, without launching a
  browser. (The strict validator IS the rule — unknown additions
  cannot silently load.)
- Given the `visual-qa` skill, When an agent needs to QA a new
  feature, Then the skill documents the schema, a working example,
  and the convention: one feature per file, one case per scenario,
  relative URLs resolved against `base_url`.

## Non-Goals

- ❌ Real-device or emulator testing (physical phones/tablets are
  outside Zed; headless emulation only).
- ❌ Cross-browser matrix — v1 drives Chrome/Chromium only.
- ❌ Pixel-diff / golden-image comparison — verdicts are LLM-based in
  v1 (a future `--diff` mode could compare two runs byte-wise; VQ-3).
- ❌ Autonomous exploration loops in v1 — the model never chooses its
  own next action (VQ-1).
- ❌ CI wiring — the tool runs locally on demand (VQ-4).
- ❌ Fixing what it finds — the tool reports; a human or another agent
  fixes.
- ❌ Running the repo's test suites — that stays QA's manual step.
- ❌ Multi-project repo restructure in v1 — single module for now
  (VQ-2).

## Follow-ups (recorded, not in v1)

- **VQ-1** Vision-driven autonomous loop: model picks the next action
  from the screenshot (needs a bounded action vocabulary + loop
  budget before it can ship).
- **VQ-2** Multi-project split: give `tools/visualqa` its own `go.mod`
  so the wiki tool's dependency tree stays rod-free.
- **VQ-3** `--diff` pixel/byte comparison between two runs.
- **VQ-4** CI wiring for the visual-qa tool.

---

## Design (Phase 2 — pending gate)

### Architecture

**Technology:** Go CLI at `tools/visualqa/` (module `deepcut-skills`,
one new dependency `github.com/go-rod/rod`). Rod launches a local
Chrome/Chromium headlessly, applies a device preset, drives actions,
captures screenshots, and exposes CDP events (console, network) for
the Diagnostics section. A thin DeepSeek vision client (OpenAI-
compatible `POST /chat/completions`) sends each screenshot with the QA
checklist and parses a structured verdict.

**Files**

| Path | Purpose |
|------|---------|
| `tools/visualqa/main.go` | CLI: flags, mode selection (one-shot vs flow), orchestration, exit codes |
| `tools/visualqa/device.go` | Device preset table (mobile/tablet/desktop: viewport, DPR, touch, UA) — pure data |
| `tools/visualqa/flow.go` | Flow schema types + strict loader (`DisallowUnknownFields`, action enum, per-case validation) |
| `tools/visualqa/vision.go` | Vision client: base64 image in a user message, JSON mode, retry/backoff, repair → re-ask → UNCERTAIN, token logging |
| `tools/visualqa/browser.go` | Rod session: launch, emulation, action executor, screenshot capture, console/network event collection |
| `tools/visualqa/report.go` | `report.md` + `findings.json` writer |
| `tools/visualqa/*_test.go` | Table-driven tests (offline; vision via `httptest` fake endpoint) |
| `.env.visualqa` | Gitignored key file at repo root (`DEEPSEEK_API_KEY=...`) |
| `.agents/skills/visual-qa/SKILL.md` | The skill (see Skill Design) |
| `wiki/`, `AGENT_INDEX.md`, `zed/profiles.json`, `wiki/catalog.json` | Registry + generated catalog updates |

`.gitignore` additions: `.env.visualqa`, `artifacts/`.

**Exit codes:** `0` = run completed and report written (FAIL findings
are findings, not tool failures); `2` = usage/flow validation error;
`3` = provider failure after retry budget; `4` = budget/timeout abort.

### CLI contract

```
go run ./tools/visualqa \
  --url <url> | --flow <feature.json> \
  --device mobile|tablet|desktop \
  [--case <name>] \
  [--checklist <text|@file>] \
  [--out artifacts/visualqa] \
  [--max-steps 15] [--max-screenshots 12] \
  [--timeout 5m] [--retries 3] [--model deepseek-v4-flash-vision-exp]
```

- `--url` and `--flow` are mutually exclusive (one-shot vs flow mode).
- One-shot mode synthesizes a single implicit case: `goto <url>` →
  `screenshot`.

### Flow schema (the authoring contract)

```json
{
  "feature": "checkout",
  "base_url": "http://localhost:3000",
  "cases": [
    {
      "name": "happy path",
      "steps": [
        { "action": "goto", "url": "/cart" },
        { "action": "screenshot", "name": "cart-empty" },
        { "action": "click", "selector": "#add-to-cart", "capture": true },
        { "action": "wait", "ms": 800 },
        { "action": "screenshot", "name": "cart-filled" },
        { "action": "scroll", "to": "bottom", "capture": true }
      ]
    },
    {
      "name": "empty cart guard",
      "steps": [
        { "action": "goto", "url": "/checkout" },
        { "action": "screenshot", "name": "guard" }
      ]
    }
  ]
}
```

**Captures are viewport-only in v1** — the vision model resizes every
image to ~800×800 total pixels, so a tall full-page capture gets
squished and loses the text detail mobile QA depends on. Below-fold
content is covered with explicit `scroll` + `screenshot` steps.
(Design refinement at the gate, 2026-08-26: US1 originally said
"full-page"; corrected to viewport after checking the resize math
against api-docs.deepseek.com.)

**Actions** (the enum; anything else is a validation error):

| action | fields | notes |
|--------|--------|-------|
| `goto` | `url` (relative → resolved against `base_url`) | navigation |
| `click` | `selector`, `capture?` | first match; step FAILS if absent |
| `type` | `selector`, `text`, `capture?` | fills an input |
| `scroll` | `to: "top"\|"bottom"` **or** `selector`, `capture?` | |
| `wait` | `ms` **or** `selector` | explicit settle |
| `screenshot` | `name` (unique per case) | always captures |

Validation rules (enforced by `flow.go`, pinned by tests):
- Top level: `feature` (non-empty), `base_url` (valid http(s) URL),
  `cases` (1..N).
- Per case: `name` (non-empty, unique within the file), `steps` (≥1).
- Per step: `action` in the enum; fields required per action;
  unknown fields rejected (`DisallowUnknownFields`); `capture` only on
  click/type/scroll.
- Relative `goto` URLs resolve against `base_url`; absolute URLs pass
  through.
- `--case <name>` selects one case; unknown name = validation error.

### Device presets

| profile | viewport | DPR | touch | UA hint |
|---------|----------|-----|-------|---------|
| `mobile` | 390×844 | 3 | yes | iPhone-class Mobile Safari |
| `tablet` | 820×1180 | 2 | yes | iPad-class Safari |
| `desktop` | 1440×900 | 1 | no | Chrome desktop |

Implemented as rod device presets; the exact viewport/DPR/touch values
are pinned in a table test (no browser needed).

### Vision contract

- **Model:** user-configurable; resolution `--model` flag >
  `DEEPSEEK_VISION_MODEL` (env, then `.env.visualqa`) > default
  `deepseek-v4-flash-vision-exp`.
- **Request:** one screenshot per call, PNG → base64 `data:` URL in a
  `user` message; image goes LAST, stable checklist prefix FIRST
  (KV-cache friendly, per ai-engineer discipline). `response_format:
  {"type": "json_object"}`, `max_tokens: 1024`.
- **Response (strictly parsed):**
  ```json
  {
    "checks": [
      { "item": "no horizontal overflow", "verdict": "PASS", "reason": "..." },
      { "item": "tap targets", "verdict": "FAIL", "reason": "bottom bar covers the primary CTA" }
    ]
  }
  ```
  `verdict ∈ {PASS, FAIL, UNCERTAIN}` (enum check), `DisallowUnknownFields`,
  item count clamped ≤ 24, reason ≤ 200 chars.
- **Reliability ladder (per call):** retry 429/500/503/network with
  backoff + full jitter honoring `Retry-After` (budget 3) → on
  malformed/empty JSON: one local repair (strip fences, first balanced
  JSON object) → one re-ask → step marked UNCERTAIN. Never re-ask on
  4xx; never retry 400/401/402/422. 402 = empty balance: fail the run
  with the provider message.
- **Cost:** ≤384 vision tokens/image → a full 12-screenshot run ≤
  ~4.6K vision tokens + text. Per-call `prompt_tokens` logged and
  summed in the report.

**Default checklist (mobile-first; overridable via `--checklist`):**
1. Content fits the viewport width — no horizontal overflow/clipping.
2. Tap targets are large enough and don't overlap.
3. Text is readable — no truncation, overlap, or poor contrast.
4. Images/media load with visible fallbacks (no broken-image icons).
5. Primary navigation and actions are reachable.
6. Layout is intentional — no blank regions or misalignment.
7. Fixed elements (headers/bottom bars) don't cover content.
8. Nothing in the frame suggests rendering breakage.

### Report format

`artifacts/visualqa/<run>/` contains `report.md`, `findings.json`, and
`s1.png`, `s2.png`, … (one per capture, named after the step).

```
# Visual QA Report — checkout
Run: 20260826T1015Z-3f2a | Device: mobile | Model: deepseek-v4-flash-vision-exp
URL/Flow: flows/checkout.json (base http://localhost:3000) | Date: ...
Budget: 5/12 screenshots, 5/15 steps | Vision tokens: 1740 | Status: COMPLETED
Summary: 6 PASS · 1 FAIL · 1 UNCERTAIN

## Findings
| # | Step | Check | Verdict | Reason |
| 1 | happy path (cart-empty) | tap targets | FAIL | "bottom bar covers the primary CTA" |

## Screenshots
![cart-empty](s1.png)
![cart-filled](s2.png)

## Diagnostics
- console error @ checkout.js:42 — "Uncaught TypeError: x is undefined"
- failed request: GET /api/cart → 404
```

`findings.json` mirrors the data machine-readably: run id, device,
url/flow, steps `[{name, screenshot, checks[]}]`, diagnostics,
summary, budget, token usage. Fail-fast aborts and budget aborts mark
the run `FAILED` with a reason in the header.

### Skill design (`visual-qa`, category: stack)

Created with the skill-factory template. `.agents/skills/visual-qa/
SKILL.md`:

- **What You Own:** running one-shot/flow visual QA, authoring flow
  JSON (schema + validator), interpreting reports.
- **What You Do NOT Own:** fixing findings, approving PRs, real-device
  testing, autonomous loops.
- **Workflow:** key setup (`.env.visualqa` at repo root) → build →
  one-shot run → flow run → authoring a new feature/case (schema +
  example + strict-validation rules + "one feature per file, one case
  per scenario") → interpreting verdicts (UNCERTAIN = re-run or human
  look; FAIL = file the finding).
- **Guardrails:** budget flags are mandatory for interactive flows;
  key only in the Authorization header (never in prompts/logs);
  screenshots stay under `artifacts/` (gitignored — may contain
  sensitive UI); no autonomous exploration (VQ-1).
- **Test Task:** author a 2-case flow for a local URL, run one case on
  mobile, produce a report, interpret it.

**Profile** (`zed/profiles.json`, new entry `visual-qa`): Flash
default / Pro heavy; tools read+grep+find+list+write+edit+terminal;
skills `["visual-qa"]`. QA sessions load the skill alongside `qa`.

**Registry updates:** `AGENT_INDEX.md` (stack-skills row),
`wiki/catalog.json` (category `stack`, tags `qa, vision, browser`),
then regenerate the wiki so `go test ./...` stays green.

### Testing approach (test-first)

Offline unit tests (no browser, no real API — no key needed):

- `flow_test.go` — schema: happy path, unknown action, unknown field,
  missing name, empty steps, duplicate case names, relative URL
  resolution, `--case` selection, unknown case name. (Writes the
  failing tests first.)
- `vision_test.go` — against an `httptest` OpenAI-compatible fake:
  success parse; JSON-mode empty content → repair; malformed → repair
  → re-ask → UNCERTAIN; 429 → retry → success; 429 exhausted → run
  fails; 400 → no retry; 402 → fail with message; token usage
  recorded.
- `device_test.go` — preset table pins (viewport/DPR/touch), no
  browser.
- `report_test.go` — `report.md` content pins (header fields, findings
  table, screenshot refs, diagnostics section) + `findings.json`
  shape.

Live path (not unit-tested; exercised by the skill's Test Task): real
Chrome launch, emulation, screenshots, one real vision round-trip —
the repo has no CI, so the browser path is verified by running it.

### Non-goal re-affirmed

No changes to the wiki tool or its stdlib-only property in this
feature (VQ-2 keeps them apart later).

## Task Checklist (Phase 3)

1. [ ] (Tooling) Add `github.com/go-rod/rod` to the module; scaffold
   `tools/visualqa/`
   → Verifies: `go build ./...` green with the wiki tool untouched
   → Satisfies: repo layout (VQ-2 deferred)

2. [ ] (Tooling) `device.go` + `device_test.go` — preset table
   → Test: `TestDevicePresets` (mobile/tablet/desktop viewport, DPR,
   touch) — red first
   → Satisfies: US3 AC1

3. [ ] (Tooling) `flow.go` + `flow_test.go` — schema + strict loader
   → Tests: happy path; unknown action; unknown field; missing name;
   empty steps; duplicate case names; relative URL resolution;
   `--case` selection; unknown case name — red first
   → Satisfies: US2 AC3, US6 AC1–AC2

4. [ ] (Tooling) `vision.go` + `vision_test.go` — client vs an
   `httptest` OpenAI-compatible fake
   → Tests: success; JSON-mode empty content → repair; malformed →
   repair → re-ask → UNCERTAIN; 429 → retry → success; 429 exhausted;
   400 → no retry; 402 → fail with message; token usage recorded;
   model resolution order (flag > env > file > default) — red first
   → Satisfies: US1 AC2, US4 AC2–AC5

5. [ ] (Tooling) `env.go` + `env_test.go` — `.env.visualqa` loader +
   precedence (process env > file)
   → Test: `TestEnvPrecedence` — red first
   → Satisfies: US5 (key convention), decision 4/6

6. [ ] (Tooling) `report.go` + `report_test.go` — `report.md` +
   `findings.json`
   → Tests: header fields; findings table; screenshot refs;
   diagnostics section; JSON shape — red first
   → Satisfies: US1 AC3, US2 AC4, US4 AC5

7. [ ] (Tooling) `browser.go` + `main.go` — rod session, action
   executor, event collection, orchestration, exit codes
   → Test: pre-browser validation paths (bad flags / bad flow exit 2
   without launching) — red first; live Chrome path exercised by the
   Test Task (repo has no CI)
   → Satisfies: US1 AC1, US2 AC1–AC2, US4 AC1

8. [ ] (Config) `.gitignore` + `.env.visualqa.example` committed;
   key/model loader wired
   → Test: `TestEnvPrecedence` (task 5)
   → Satisfies: decision 4/6, US5

9. [ ] (Skill) `.agents/skills/visual-qa/SKILL.md` + `zed/profiles.json`
   entry + `AGENT_INDEX.md` row + `wiki/catalog.json` entry;
   regenerate the wiki
   → Verifies: ruby YAML parse; `python3 -m json.tool`;
   `go test ./...` green (catalog + wiki up-to-date checks)
   → Satisfies: US5 AC1–AC2, US6 AC3

10. [ ] (Live) Test Task: one-shot + a 2-case flow against a local
    page on mobile; live vision round-trip with the real key
    → Run by the operator (key lives in the gitignored `.env.visualqa`)
    → Satisfies: US1/US2 end-to-end
