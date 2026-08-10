# HOW WE WORK — Multi-Role AI Development

## The Setup

You run six agent threads simultaneously in Zed, each with a different
role skill loaded:

| Thread | Skill | Can Write? | Can Terminal? | Model (fast) | Model (heavy) |
|--------|-------|-----------|--------------|-------------|-------------|
| PM | `pm` | specs only | ❌ | DeepSeek-V4 | — |
| Architect | `architect` | specs only | ❌ | DeepSeek-V4 | Claude Opus* |
| UX Designer | `ux-designer` | specs only | ❌ | DeepSeek-V4 | Claude Opus* |
| Backend Eng | `backend-engineer` | `backend/` | ✅ | DeepSeek-V4 | Claude Opus* |
| Frontend Eng | `frontend-engineer` | `frontend/` | ✅ | DeepSeek-V4 | — |
| Mobile Eng | `mobile-engineer` | `mobile/` | ✅ | DeepSeek-V4 | — |
| Reviewer | `reviewer` | ❌ | ✅ | DeepSeek-V4 | Claude Opus* |
| QA | `qa` | ❌ | ✅ | DeepSeek-V4 | — |
| Security Eng | `security-engineer` | ❌ | ✅ | DeepSeek-V4 | Claude Opus* |

*Use stronger model for architecture decisions, complex type design,
cross-cutting refactors, and security audits. DeepSeek is fine for
routine implementation.

---

## The Specs Directory

```
specs/
  <feature-slug>.md     # The living spec for one feature
```

`specs/` is **the single source of truth**. Every role reads from it,
writes to it, or builds against it. It is version-controlled and
reviewed.

A spec file contains, in order:

1. **Metadata**: feature name, status (Draft/Review/Approved/Implemented/Archived), owner
2. **Requirements**: user stories + Given/When/Then acceptance criteria
3. **Explicit Non-Goals**: what we are NOT building
4. **Design**: architecture (Architect) + API contract (Architect) + data model (Architect) + UI design (UX Designer)
5. **Task Checklist**: ordered, independently-reviewable items
6. **Implementation Notes**: decisions made during build, deviations from spec

---

## Feature Lifecycle (Gated Phases)

```
                        HUMAN CHECKPOINT
                             │
  ┌──────────┐    ┌──────────┼──────────┐    ┌──────────────────────────────────┐
  │ Phase 1  │───▶│  Review  │          │───▶│  Phase 2: Design                 │
  │ Req'ts   │    │  Gate    │          │    │  Architect + UX Designer         │
  └──────────┘    └──────────┘          │    │  (parallel threads, shared spec) │
        PM owns                         │    └──────────────────────────────────┘
                                        │                    │
                              ┌─────────┼──────────┐         │ HUMAN CHECKPOINT
                              │  Review  │          │         │
                              │  Gate    │          │         │
                              └──────────┘          │         ▼
                                                    │    ┌────────────────┐
                                                    │    │  Phase 3       │
                                                    │    │  Task Breakdown│
                                                    │    └────────────────┘
                                                    │          PM owns
                                                    │               │
                                                    │    HUMAN CHECKPOINT
                                                    │               │
                                                    ▼               ▼
                                             ┌────────────────────────────┐
                                             │  Phase 4: Implementation   │
                                             │  (parallel across engineers)│
                                             └────────────────────────────┘
```

### Phase 1: Requirements (PM thread)

1. Load `pm` skill.
2. Read any existing context (user prompt, product brief).
3. Write `specs/<feature>.md` with sections: Metadata, Requirements,
   Non-Goals.
4. Present to user at the **Review Gate**.
5. User approves or requests changes. Iterate in the PM thread until
   approved and status is set to `Approved`.

### Phase 2: Design (Architect + UX Designer threads, parallel)

**Architect:**
1. Load `architect` skill.
2. Read the approved Requirements from `specs/<feature>.md`.
3. Add Architecture section: system design, API contract, data model.
4. Coordinate with UX Designer on the API/UI boundary.

**UX Designer:**
1. Load `ux-designer` skill.
2. Read the approved Requirements + the Architect's API contract.
3. Add UI Design section: component inventory, states, design tokens,
   accessibility, user flows, UX copy.
4. Coordinate with Architect if the UI needs data the API doesn't provide.

Both roles write to the same `specs/<feature>.md`. Present together at
the **Review Gate**. User approves. Status moves to `Approved`.

### Phase 3: Task Breakdown (PM thread)

1. PM reads the approved Design (both Architecture and UI Design).
2. Adds Task Checklist: ordered items, each independently reviewable.
3. Each task includes: files to touch, acceptance criteria to verify,
   which role owns it.
4. No human gate here — tasks are derived mechanically from the design.
   Status moves to `Approved`.

### Phase 4: Implementation (Engineer threads)

1. Backend Engineer starts first if the feature has an API component.
   - Publishes stable API contract before Frontend/Mobile build against it.
2. Frontend and Mobile Engineers can work in parallel once the API
   contract is stable. They build against the UX Designer's component
   specs and design tokens.
3. Each engineer works through the task checklist, one task at a time.
4. After each task, verify against the acceptance criteria.
5. Mark tasks `[x]` in the spec as completed.
6. When all tasks are done, status moves to `Implemented`.

---

## Handoff Protocol

### PM → Architect + UX Designer

PM writes: "Architect and UX Designer, `specs/<feature>.md` Requirements
section is approved and stable. Architect: produce API contract + data
model. UX Designer: produce UI design. Coordinate on the boundary."

### Architect → UX Designer

Architect writes: "API returns these shapes: [summary]. Data model:
[tables]. UX Designer, build UI layer against this."

### UX Designer → Architect

UX Designer writes: "Component X needs field Y that isn't in the API
response. Can we add it to the contract?"

### Architect → Engineers

Architect writes: "API contract and data model for `<feature>` are in
`specs/<feature>.md` Design section. This contract is stable — build
against it."

### UX Designer → Engineers

UX Designer writes: "Component designs, states, tokens for `<feature>`
are in `specs/<feature>.md` UI Design section. Build against these."

### Engineer → Engineer

- Backend Engineer: "Route `GET /api/<resource>` now returns the
  contract shape. See `specs/<feature>.md` for the spec."
- Frontend/Mobile: "Acknowledged. Building UI against that contract."

### Any Role → PM (ambiguity)

If any role encounters ambiguity in the spec:
1. Do NOT guess.
2. Open a comment in the spec file with `[NEEDS CLARIFICATION: ...]`.
3. Tag the PM thread.
4. PM resolves and updates the spec.
5. Resume work from the updated spec.

---

## Parallel vs. Sequential Execution

### Always Sequential (gated)

- Requirements → Design → Implementation
- API contract design → Backend implementation → Frontend/Mobile
  consumption

### Parallel (same phase)

- Architect and UX Designer work on the same spec simultaneously
- Backend Engineer can implement multiple independent endpoints
- Frontend and Mobile Engineers work on their UI layers in parallel
  once both the API contract and UI design are stable
- PM, Architect, and UX Designer can plan the NEXT feature while
  engineers implement the current one
- QA and Security Engineer audit the PR in parallel before final approval

---

## Model Selection as a Cost Lever

| Task Complexity | Model | Cost |
|-----------------|-------|------|
| Routine CRUD, wiring, boilerplate | DeepSeek-V4 | Low |
| Architecture decisions, type design, complex logic | Claude Opus / GPT-4o | Medium |
| Spec writing, acceptance criteria | DeepSeek-V4 | Low |
| UI design, design system work | DeepSeek-V4 / Claude Opus | Low/Medium |
| Cross-cutting refactors, security audits | Claude Opus | Medium |

Rule of thumb: use DeepSeek for 80% of work. Escalate to a stronger
model when DeepSeek gets stuck, hallucinates, or when the decision is
hard to reverse.

---

## Iterating on Skills

After each feature, do a brief retrospective:

1. What did the agent do that a rule should have prevented?
2. What rule was missing?
3. What rule was too vague to be useful?
4. Update the relevant SKILL.md or AGENTS.md.

Keep a `specs/memories/` directory for these retrospectives:
```
specs/memories/2026-08-06-feature-x-retro.md
```

---

*Last updated: 2026-08-06*
