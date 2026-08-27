# Visual QA: Browser Automation + DeepSeek Vision

**Status:** Draft
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
  when fine detail doesn't matter).

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
  URL, waits for load, and captures a full-page screenshot saved under
  `artifacts/visualqa/<run>/`.
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
- Given a flow file (`--flow flows/<name>.json`), When the tool runs,
  Then it executes the steps in order and captures a screenshot at
  every step marked `"capture": true`.
- Given a step that fails to locate its target element, When the tool
  runs, Then it records the step as FAILED with the CDP error, aborts
  the remaining steps (fail-fast), and includes the failure in the
  report.
- Given a flow file with an unknown action or malformed JSON, When the
  tool runs, Then it exits non-zero with a validation error naming the
  step, **without** launching a browser.

### User Story 3: Desktop / tablet / mobile presets

As a QA engineer, I want the three form factors available from one
flag so the same flow runs across device classes.

**Acceptance Criteria:**
- Given `--device desktop|tablet|mobile`, When the tool runs, Then the
  browser emulates the corresponding profile (mobile: phone viewport +
  touch + UA; tablet: tablet viewport + touch; desktop: default
  desktop viewport) and the report records the profile used.
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

## Non-Goals

- ❌ Real-device or emulator testing (physical phones/tablets are
  outside Zed; headless emulation only).
- ❌ Cross-browser matrix — v1 drives Chrome/Chromium only.
- ❌ Pixel-diff / golden-image comparison — verdicts are LLM-based in
  v1 (a future `--diff` mode could compare two runs byte-wise).
- ❌ Autonomous exploration loops in v1 — the model never chooses its
  own next action (see Open Questions; candidate for v2).
- ❌ CI wiring — the tool runs locally on demand; CI integration is a
  follow-up.
- ❌ Fixing what it finds — the tool reports; a human or another agent
  fixes.
- ❌ Running the repo's test suites — that stays QA's manual step.

## Open Questions

- [NEEDS CLARIFICATION: interaction model — v1 ships scripted flows
  with per-step vision verification. Do you also want the
  vision-driven loop (model picks the next action from the
  screenshot) in v1, or as a follow-up?]
- [NEEDS CLARIFICATION: the tool needs a Go browser-automation
  dependency (recommended `github.com/go-rod/rod`; alternative
  `github.com/chromedp/chromedp`) — the repo is currently
  stdlib-only. OK to add one?]
- [NEEDS CLARIFICATION: default budget numbers — max-steps 15 /
  max-screenshots 12 / 5m timeout / 3 retries. Acceptable defaults?]
- [NEEDS CLARIFICATION: API key — `DEEPSEEK_API_KEY` read from a
  gitignored `.env.visualqa` in the repo root (also honors the
  `DEEPSEEK_API_KEY` environment variable). OK?]
