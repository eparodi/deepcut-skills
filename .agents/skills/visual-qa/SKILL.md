---
name: visual-qa
description: "Visual QA — drives headless Chrome via tools/visualqa with DeepSeek vision: one-shot page verification, multi-case flow JSON authoring, and report interpretation across mobile/tablet/desktop viewports. Load when a QA session needs to verify rendered UI states or author visual-QA flows."
---

# Visual QA

**Profile:** `visual-qa` in `zed/profiles.json` — model tier lives in
the profile, never here.

You run automated visual QA: a headless Chrome session captures
viewport screenshots at scripted points, the DeepSeek vision model
(`deepseek-v4-flash-vision-exp` by default) grades each against a QA
checklist, and the tool writes a report. Spec: `specs/visual-qa.md`.

## What You Own

- Running visual QA with `go run ./tools/visualqa` (one-shot and flow
  modes)
- Authoring flow JSON — one feature per file, one case per scenario
- Interpreting `report.md` / `findings.json` verdicts

## What You Do NOT Own

- ❌ Fixing findings — you report them; a human or another agent fixes
- ❌ Approving PRs or deciding "done"
- ❌ Real-device or emulator testing (headless emulation only)
- ❌ Autonomous exploration loops (VQ-1 — the model never picks its
  own next action)

## Prerequisites

- Chrome/Chromium installed locally (rod launches it headlessly)
- API key in the gitignored `.env.visualqa` at the repo root
  (`DEEPSEEK_API_KEY=...`), or the `DEEPSEEK_API_KEY` env var
- One-time setup: `go mod download` (rod is already in `go.mod`)

## Workflow

### 1. One-shot verification

```bash
go run ./tools/visualqa --url <url> --device mobile
```

Read the report at `artifacts/visualqa/<run>/report.md`.

### 2. Flow verification

Flow files live in `flows/` at the repo root:

```bash
go run ./tools/visualqa --flow flows/checkout.json --device mobile
go run ./tools/visualqa --flow flows/checkout.json --case "empty cart guard" --device desktop
```

### 3. Authoring a new feature or case (the JSON contract)

One feature per file; one case per scenario; relative `url`s resolve
against `base_url`:

```json
{ "feature": "checkout",
  "base_url": "http://localhost:3000",
  "cases": [
    { "name": "happy path",
      "steps": [
        { "action": "goto", "url": "/cart" },
        { "action": "screenshot", "name": "cart-empty", "mode": "full", "html": true },
        { "action": "click", "selector": "#add-to-cart", "capture": true },
        { "action": "scroll", "to": "bottom", "capture": true },
        { "action": "screenshot", "name": "cart-filled" }
      ] },
    { "name": "empty cart guard",
      "steps": [
        { "action": "goto", "url": "/checkout" },
        { "action": "screenshot", "name": "guard" }
      ] } ] }
```

Actions enum: `goto` (`url`), `click` (`selector`), `type`
(`selector`, `text`), `scroll` (`to: "top"|"bottom"` **or**
`selector`), `wait` (`ms` **or** `selector`), `screenshot` (`name`).
`"capture": true` on click/type/scroll takes a viewport shot after the
action. Captures are viewport-only by default — below-fold content
needs explicit `scroll` + `screenshot` steps (tall full-page captures
get squished by the vision model's ~800×800 resize and lose text
detail).

The loader is strict: unknown actions/fields, missing names,
duplicate names, or empty steps exit 2 naming the case/step BEFORE a
browser launches. Run the flow to confirm a new case parses.

### 3c. Capture modes & HTML evidence

Every capture-capable step (`screenshot`, and click/type/scroll with
`capture: true`) takes two optional knobs:

- `"mode": "viewport"|"full"` — `viewport` (default) captures the
  visible frame; `full` captures the whole scrollable page in one
  image (rod's full-page capture; fixed elements render once at the
  top). Run-level default: `--capture-mode viewport|full`.
- `"html": true|false` — sends the sanitized page HTML (scripts,
  styles, comments stripped; capped at `--max-html-chars`, default
  30k chars ≈ 7–8k tokens, marked `…[truncated]`) as a text block
  alongside the image, so the model verifies structural evidence the
  downscaled image can't show (element sizes, labels, aria, alt,
  hrefs). Run-level default: `--with-html`.

Per-step values override run-level defaults. Full-page captures are
downscaled to fit the vision API's 8192 px-per-side limit before
sending.

**Capture mode changes findings** — proven live (2026-08-27, bot
dashboards): the same flow in viewport vs full+HTML moved verdicts
from 68/7/37 to 49/9/27 (PASS/FAIL/UNCERTAIN), with a different FAIL
set per page. The HTML channel also raises cost to ~9k tokens/step.
Pick the mode per step based on what the finding must answer:
viewport shows what the user sees; full+HTML shows the whole page
with structural precision.

### 3d. Authenticated exploration (cookie + autonomous loop)

Two knobs unlock pages behind the login and let the model drive the
browser on its own:

```bash
# inject a session cookie (flag > VISUALQA_COOKIE env > .env.visualqa)
# then explore: the model grades each page AND picks the next action
# (click / scroll / back / goto / done; + type with --test-env)
go run ./tools/visualqa --explore --url http://app/ \
  --cookie "session=<signed-token>" --device mobile
```

- The bot dashboard's stateless sessions are a `session` cookie with
  an HMAC-SHA256-signed token (`1.<b64url claims>.<b64url sig>`, 24h
  TTL). The OPERATOR mints it (their tooling + the server's
  `SESSION_SECRET`); the tool only injects whatever `name=value` it is
  given. The value stays in the CDP session — never in reports/logs.
- Exploration is same-origin only; the identical action twice in a row
  ends the loop (anti-loop guard); the same step/screenshot/timeout
  caps bound it; HTML is always included (~9k tokens/step, quick
  8-item checklist by default).
- `--test-env` unlocks `type` (form-filling, state mutation) — for
  disposable test instances only, never live state.
- Live-validated 2026-08-28: injected cookie → authed dashboard →
  model clicked a details link, scrolled, clicked the trades nav, then
  said done: 31 PASS / 0 FAIL / 1 UNCERTAIN over 3 pages.

### 3b. Choosing a checklist (the library)

The default 8-item quick checklist is fine for a first pass. For
thorough per-device QA, `tools/visualqa/checklists/` holds curated,
source-backed checklists (Google web.dev, W3C/WCAG 2.2, NN/g,
Smashing, CSS-Tricks, MDN, LukeW, A List Apart, Baymard — see
`docs/research/visual-qa-device-design-sources.md`) grouped by type:

```bash
# one group on mobile (the primary design target)
go run ./tools/visualqa --url <url> --device mobile \
  --checklist @tools/visualqa/checklists/mobile/layout.md

# ALL groups: pass the directory — every .md file inside is read,
# sorted by filename, and concatenated (no all.md file to maintain)
--checklist @tools/visualqa/checklists/mobile/
```

Groups: `layout`, `typography`, `navigation`, `forms`, `media`,
`interactions`, `accessibility` (one file per device × group). Each
device is tuned to its form factor — mobile is touch/thumb/safe-area
centric, tablet adds hybrid hover+touch and split-pane, desktop adds
keyboard operation and window-resize behavior. A full device run is
28 rules, under the tool's 32-check response clamp.

**Rule precision:** every rule states an exact bar (44×44 CSS px touch
targets, WCAG 2.5.8's 24×24, 4.5:1/3:1 contrast, 60-80ch line
length), the in-frame evidence that supports it, and when the frame
cannot show it. The model is instructed to mark UNCERTAIN when the
evidence is not visible — never PASS on missing evidence. An
UNCERTAIN verdict means "needs a human or a different check", not a
weak pass.

### 4. Interpreting verdicts

- **PASS** — the item clearly holds.
- **FAIL** — the item is clearly broken; the report's reason +
  screenshot are the finding.
- **UNCERTAIN** — the screenshot couldn't tell (or the model's answer
  was unparseable); re-run or take a human look.
- **Run Status** — `COMPLETED` (findings may still exist) vs `FAILED`
  (aborted: budget, timeout, step error, provider failure).
- The **Diagnostics** section lists console errors and failed network
  requests captured from CDP events — worth checking before trusting
  a PASS.

## Interpreting verdicts — context matters

A FAIL is only a finding if the page SHOULD have the thing. On a
pre-auth login page, nav-reachability/back-check FAILs are expected
context (no nav exists yet) — do not file them as regressions.
Check the run's screenshots and step order before trusting a FAIL.

## Guardrails

- Budget flags are mandatory for interactive flows: `--max-steps`,
  `--max-screenshots`, `--timeout`, `--retries` (defaults
  15 / 12 / 5m / 3). A run that exceeds a cap aborts and marks the
  run FAILED.
- Capture/HTML flags: `--capture-mode viewport|full` (default
  viewport), `--with-html`, `--max-html-chars` (default 30k). HTML
  raises per-step cost to ~9k tokens — budget `--timeout` and retries
  accordingly.
- Auth/explore flags: `--cookie "name=value"`, `--explore` (requires
  `--url`), `--test-env` (unlocks `type` — test instances only).
- The vision API hard-rejects images with a side > 8192px; full-page
  captures are downscaled to fit automatically.
- rod's `MustEval` wraps every script as `(%s).apply(this, arguments)`
  — write JS as function expressions (`function() { ... }`), never
  bare statements (`Math.max(...)` panics with `.apply is not a
  function`).
- The API key goes ONLY in the Authorization header — never in
  prompts, logs, or reports.
- Screenshots may contain sensitive UI — they stay under the
  gitignored `artifacts/` directory.
- Model is user-configurable: `--model` flag > `DEEPSEEK_VISION_MODEL`
  (env, then `.env.visualqa`) > default `deepseek-v4-flash-vision-exp`.
- No autonomous exploration loops in v1 (VQ-1).
- Retries happen on 429/500/503 and on 200s with unparseable
  (truncated) bodies — a 4xx (including 402 empty balance) fails the
  run with the provider message — never re-send a malformed payload.

## Test Task

Author a 2-case flow for a local page (a local dev server is fine),
run one case on mobile, produce the report, and interpret it — one
PASS/FAIL/UNCERTAIN per checklist item. Confirm the strict loader
rejects an unknown action with exit 2.

## Handoff

Report: run id, device, status, per-step verdicts, diagnostics
(console errors / failed requests), and screenshot paths. Output
`[VISUAL_QA_DONE]` when complete.
