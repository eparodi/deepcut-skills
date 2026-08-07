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
   - Which acceptance criteria it satisfies
   - Which role owns it
4. Mark tasks that can run in parallel with `[P]`:
   ```markdown
   ## Task Checklist
   
   1. [ ] (Backend) Create `users` migration
      → Satisfies: data model
   
   2. [ ] (Backend) Implement `POST /api/users` handler
      → Satisfies: US1 AC1, US1 AC2
   
   3. [ ] [P] (Frontend) Add UserForm component
      → Satisfies: US1 AC3
   
   4. [ ] [P] (Frontend) Add user list page
      → Satisfies: US1 AC4
   
   5. [ ] [P] (Mobile) Add UserForm screen
      → Satisfies: US1 AC3
   ```

## Phase 4 — Implementation

**Owner:** Engineer roles (Backend, Frontend, Mobile).  
**Input:** Approved spec with task checklist.

### Process

1. Load the relevant stack-specific skill (`go-chi`, `nextjs`, `expo`).
2. Work through tasks ONE AT A TIME, in order.
3. For each task:
   a. Read the relevant section of the spec.
   b. Implement the change.
   c. Run the build. If it fails, fix it.
   d. Run targeted tests. If they fail, fix them.
   e. Verify against acceptance criteria.
   f. Mark the task `[x]` in the checklist.
4. When all tasks are done, set status to `Implemented`.

### Rules During Implementation

- **Backend first.** If a task requires a new API endpoint, the Backend
  Engineer must implement and stabilize it before Frontend/Mobile build
  against it.
- **Do not change the spec.** If you discover the spec is wrong during
  implementation, add a note under `## Implementation Notes` and continue
  working to the spec. Flag the issue for the PM. Do NOT silently
  deviate from the spec.
- **Commit per task.** Each completed task gets its own commit with a
  descriptive message.
- **Stop at test failures.** Do not move to the next task until the
  current one passes all tests.

---

## Quick Reference

| Phase | Who | Output | Gate? |
|-------|-----|--------|-------|
| 1. Requirements | PM | User stories + AC + Non-Goals | ✅ Human |
| 2. Design | Architect + UX Designer | Architecture + API + Data Model + UI Design | ✅ Human |
| 3. Task Breakdown | PM | Ordered checklist with role assignments | ❌ Auto |
| 4. Implementation | Engineers | Working code | ✅ Tests |
