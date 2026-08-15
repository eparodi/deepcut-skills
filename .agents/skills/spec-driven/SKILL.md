---
name: spec-driven
description: Design and implement features using a specification-first approach with gated phases. Draft requirements, get approval, design architecture, get approval, break down tasks, then implement one task at a time with verification against acceptance criteria.
---

# Spec-Driven Development

Take every non-trivial task through four gated phases. The goal is to
catch misunderstandings in cheap markdown edits instead of expensive
code rewrites. Each phase ends with a **human checkpoint** before the
next phase begins.

## Phase 0 — Triage: "Is This Non-Trivial?"

Before entering the full workflow, decide if this task even needs it.

**Use spec-driven when:**
- The task touches more than one file or one layer
- There are multiple user-facing behaviors to define
- The data model or API contract will change
- Another role (frontend, mobile) will consume what you build

**Skip spec-driven when:**
- The task is a single-line bug fix
- The change is purely mechanical (rename, extract function)
- The user explicitly says "just do it"

If unsure, ask the user: "This could benefit from a spec-driven approach — want me to draft requirements first, or should I just make the change?"

## Phase 1 — Requirements

**Owner:** PM role.  
**Output:** `specs/<feature-slug>.md` with Metadata, Requirements, and Non-Goals sections.

### Process

1. Read any existing context (user prompt, product brief, related specs).
2. Draft user stories in the format:
   ```
   As a <role>, I want <capability> so that <benefit>.
   ```
3. For each story, write Given/When/Then acceptance criteria:
   ```
   Given <precondition>
   When <action>
   Then <expected outcome>
   ```
   Every acceptance criterion must be observable from the outside (HTTP
   response, rendered UI state, file/DB effect) — if it can't be turned
   into an integration test, it's not an acceptance criterion.
4. List explicit non-goals — things we are deliberately NOT building:
   ```
   ## Non-Goals
   - ❌ Password reset (out of scope for v1)
   - ❌ Admin dashboard (handled by separate feature)
   - ❌ Email notifications (will be a separate spec)
   ```
5. Mark all ambiguities with `[NEEDS CLARIFICATION: question]`. Do NOT guess.
6. Format the spec using this template:
   ```markdown
   # <Feature Name>
   
   **Status:** Draft  
   **Owner:** <name>  
   **Created:** <date>
   
   ## Requirements
   
   ### User Story 1: <title>
   As a <role>, I want <capability> so that <benefit>.
   
   **Acceptance Criteria:**
   - Given <precondition>, When <action>, Then <expected>
   
   ### User Story 2: ...
   
   ## Non-Goals
   - ❌ <explicitly out of scope>
   ```

### Review Gate

Present the spec to the user:
```
Requirements draft is ready at specs/<feature-slug>.md

Summary:
- N user stories with acceptance criteria
- Key non-goals: [list]
- Open questions: [list any NEEDS CLARIFICATION markers]

Review and approve to proceed to Design, or request changes.
```

## Phase 2 — Design

**Owners:** Architect + UX Designer (parallel threads).  
**Output:** Add Architecture section (Architect) and UI Design section (UX Designer) to the spec.

### Architect's Process

1. Read the approved Requirements from the spec.
2. Design the system architecture:
   - What new routes/endpoints? What existing ones change?
   - What new database tables/columns? What migrations?
   - Technology decisions with rationale
3. Define the API contract:
   ```markdown
   ## API Contract
   
   ### POST /api/users
   **Request:**
   ```json
   {
     "name": "string, required, 1-100 chars",
     "email": "string, required, valid email"
   }
   ```
   
   **Response 201:**
   ```json
   {
     "id": "uuid",
     "name": "string",
     "email": "string",
     "createdAt": "ISO8601"
   }
   ```
   
   **Error Responses:**
   - 400: Validation error
   - 409: Email already exists
   ```
4. Define the data model:
   ```sql
   CREATE TABLE users (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
     email TEXT NOT NULL UNIQUE,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   ```
5. State design decisions with rationale.

### UX Designer's Process

1. Read the approved Requirements + the Architect's API contract.
2. Design the UI layer:
   - Component inventory with variants and states
   - Design tokens (colors, typography, spacing)
   - User flows for key interactions
   - Accessibility requirements
   - UX copy
3. Coordinate with Architect on the API/UI boundary.

### Review Gate

Present the design to the user. Both Architect and UX Designer sections
should be complete before the review gate. Do NOT proceed until approved.
If the API contract changes during implementation, route back through
the Architect — do NOT silently change the contract.

## Phase 3 — Task Breakdown

**Owner:** PM role.  
**Output:** Add Task Checklist section to the spec.

### Process

1. Read the approved Design (both Architecture and UI Design).
2. Break the implementation into ordered, independently-reviewable tasks.
3. Each task includes:
   - What files to create/edit
   - What the task delivers (concrete, verifiable)
   - **Which test will be written first** (see Phase 4 test-first rules)
   - Which acceptance criteria it satisfies
   - Which role owns it
4. Mark tasks that can run in parallel with `[P]`:
   ```markdown
   ## Task Checklist

   1. [ ] (Backend) Create `users` migration
      → Test: backend/cmd/server/server_test.go::TestCreateUser
      → Satisfies: data model

   2. [ ] (Backend) Implement `POST /api/users` handler
      → Test: backend/cmd/server/server_test.go::TestCreateUser
      → Satisfies: US1 AC1, US1 AC2

   3. [ ] [P] (Frontend) Add UserForm component
      → Test: frontend/src/components/UserForm.test.tsx
      → Satisfies: US1 AC3

   4. [ ] [P] (Frontend) Add user list page
      → Test: frontend/src/app/users/page.test.tsx
      → Satisfies: US1 AC4

   5. [ ] [P] (Mobile) Add UserForm screen
      → Test: mobile/__tests__/UserForm.test.tsx
      → Satisfies: US1 AC3
   ```

## Phase 4 — Implementation (Test-First)

**Owner:** Engineer roles (Backend, Frontend, Mobile).  
**Input:** Approved spec with task checklist.

### Process

1. Load the relevant stack-specific skill (`go-chi`, `nextjs`, `expo`).
2. Work through tasks ONE AT A TIME, in order.
3. For each task:
   a. Read the relevant section of the spec.
   b. **Write the failing test first (Red).** Create the integration test
      named in the task checklist BEFORE the implementation. It encodes
      the acceptance criteria: happy path first, then each error path.
   c. **Run the test and confirm it fails** — if it passes before the
      implementation exists, it proves nothing about the new code.
      A compile error counts as red; prefer the test to fail for the
      right reason (missing route/behavior) once it compiles.
   d. Implement the change (Green).
   e. Run the build. If it fails, fix it.
   f. Run targeted tests. If they fail, fix them.
   g. Verify against acceptance criteria.
   h. Mark the task `[x]` in the checklist.
4. When all tasks are done, set status to `Implemented`.

### Test-First Rules

- **Every happy path gets an integration test before implementation.**
  The test must exercise real infrastructure — `httptest.NewServer` +
  testcontainers Postgres (Go), or the docker compose stack for
  media/external-service flows — not mocks alone. Mocks may accompany
  it, but the integration test is the contract.
- **Every bug found during the feature gets a regression test first.**
  Write the test that reproduces the bug (confirm it fails), then fix.
  This applies to bugs found in review, QA, or user testing.
- **Each error path and edge case gets its own test** (table-driven where
  the stack skill prescribes it), written alongside the happy path.
- **When to skip test-first:** pure refactors with no observable
  behavior change, config-only changes, and one-line fixes without
  observable behavior — these are test-after (still verify the suite
  stays green).
- **A passing test proves nothing about new code.** Always watch the new
  test fail before writing the implementation.
- **Red commits must not be the last thing pushed.** A commit containing
  only the failing test intentionally fails CI. Push it together with
  (or after) its green counterpart so the branch HEAD is always green.

### Rules During Implementation

- **Backend first.** If a task requires a new API endpoint, the Backend
  Engineer must implement and stabilize it before Frontend/Mobile build
  against it. (The backend's integration test is written and failing
  before the endpoint exists — red first still applies.)
- **Do not change the spec.** If you discover the spec is wrong during
  implementation, add a note under `## Implementation Notes` and continue
  working to the spec. Flag the issue for the PM. Do NOT silently
  deviate from the spec.
- **Stale spec conditions.** Spec instructions with conditions ("remove
  field X until the backend implements it") go stale when another feature
  implements the missing piece. When the condition flips, keep the
  current contract and log an Implementation Note instead of blindly
  following the stale instruction.
- **Commit per task.** Each completed task gets its own commit with a
  descriptive message. The test and its implementation go in the same
  commit (or two: test commit followed by implementation commit).
- **Stop at test failures.** Do not move to the next task until the
  current one passes all tests.

---

## Quick Reference

| Phase | Who | Output | Gate? |
|-------|-----|--------|-------|
| 1. Requirements | PM | User stories + AC + Non-Goals | ✅ Human |
| 2. Design | Architect + UX Designer | Architecture + API + Data Model + UI Design | ✅ Human |
| 3. Task Breakdown | PM | Ordered checklist with role assignments + `→ Test:` per task | ❌ Auto |
| 4. Implementation | Engineers | Working code + tests written first (red → green) | ✅ Tests |

---

---

## Prerequisites (for Post-Approval Pipeline)

- `gh` (GitHub CLI, `gh auth login` done) — PR creation and CI polling.
- CI status is polled with `gh pr checks <pr>`; no inline scripts.

## Post-Approval Orchestration (Automatic Pipeline)

> Activates immediately after the spec reaches Approved status. Drive
> the implementation without asking for product or technical decisions;
> pause only on genuine spec ambiguity.

**Branch:** `feat/<slug>` from latest `main`, pushed early.
**Permission override:** inside this pipeline you may commit and push
freely (the usual "never commit without explicit request" rule is
suspended); you may NOT merge — only the user merges.

Phases:

1. **Task breakdown** — read the approved spec; break it into
   implementation tasks grouped by layer in dependency order (backend
   contracts before frontend/mobile consumption). Print the task list.
2. **Implementation** — for each layer: load the role's skill, write
   the failing tests first (happy paths, then error paths), implement
   until green, run the layer's suite + lint, commit + push. Output
   the completion marker `[<LAYER>_COMPLETE]`.
3. **PR** — once all layers are complete:
   `gh pr create --base main --head feat/<slug> --draft`.
4. **CI gate** — poll `gh pr checks <pr>` (10 min timeout). FAIL →
   fix, push, re-poll. TIMEOUT → tell the user "CI still running; say
   **next** when checks are green."
5. **Reviewer** — load the `reviewer` skill; severity ≥ major issues
   → fix and loop back to CI. Clean → proceed.
6. **QA + Security (parallel)** — both must output `[QA_PASS]` and
   `[SECURITY_PASS]`; failures loop back to CI.
7. **Ready for user** — all gates green → final message with PR
   number, spec, branch, and gate status. Do not merge.

**Failing loops:** if the same issue fails 3 times, stop and ask the
user for guidance.

The single-thread alternative (one agent, no PR ceremony) is the
`orchestrator` skill — one pipeline, do not duplicate these phases.
