# AGENTS.md — Always-On AI Agent Rules

> **Loaded automatically for every agent thread in this project.**
> These rules address DeepSeek-specific failure modes and enforce
> minimum quality bars. Skill files add stack-specific and role-specific
> rules on top of these.

---

## Section 1 — DeepSeek-Specific Guardrails

DeepSeek is powerful but has known failure modes. These rules target them.

### 1.1 Hallucinate-First Prevention

DeepSeek tends to fabricate imports, APIs, and function signatures more
often than other models. Counter with:

- **Cite your source.** Before using any library API, grep the actual
  import path used in this codebase. If it doesn't exist, do NOT guess.
  ```bash
  grep -r "import.*<package>" --include="*.go" backend/
  grep -r "from.*<package>" --include="*.ts*" frontend/
  ```
- **"I don't know" is a valid answer.** If you cannot find the exact
  API you need in the codebase or in known documentation, say so and
  ask the user. Do NOT fabricate.
- **Verify before writing.** When adding a new import or function call,
  run the build immediately after to catch wrong signatures.
- **Prefer stdlib and existing deps.** Never introduce a new dependency
  without explicit user approval. If the codebase already has a pattern
  for something, follow it rather than inventing a new one.

### 1.2 Planning Deficiency

DeepSeek tends to skip upfront design and jump to code. Force a plan:

- **Non-trivial tasks require a spec.** If the user asks for a feature
  that spans more than one file or one layer, route to the spec-driven
  skill (`@spec-driven`) first.
- **State your approach before editing.** Before touching files,
  describe: (1) what you'll change, (2) which files, (3) what the
  change accomplishes. The user can correct misalignment before you
  touch code.
- **One logical change per commit.** Do not bundle refactors with
  features, or bugfixes with enhancements.
- **Ambiguous verbs default to read-only.** If the user's instruction
  uses an inspection verb ("check", "look at", "review", "what's wrong",
  "show me"), default to reporting findings. Do NOT fix, commit, or push
  unless the user explicitly asks. When in doubt, ask: "I found X. Want
  me to fix it?"

### 1.3 Over-Confidence Calibration

DeepSeek can sound very sure about incorrect things.

- **Flag uncertainty explicitly.** If you're inferring behavior or
  making an assumption, label it: `[Inference]`, `[Assumption]`,
  `[Unverified]`.
- **Test every claim.** If you say "this should work", run it and
  prove it. Never end a turn with untested code and a statement of
  confidence.
- **When stuck, narrow the problem.** Isolate the failing piece before
  attempting broader fixes.

---

## Section 2 — Codebase Pattern Matching

### 2.1 Match Existing Patterns, Do Not Invent

- **File structure.** New files go where equivalent files already live.
  If handlers live in `backend/internal/handler/`, put yours there too.
  Do not create `backend/handlers/` because a different framework does.
- **Naming conventions.** Match the project's existing naming (camelCase,
  PascalCase, snake_case, kebab-case) exactly. Check adjacent files.
- **Error handling style.** Match the wrapping pattern (e.g.,
  `fmt.Errorf("context: %w", err)` vs `errors.Wrap`).
- **Testing style.** Match table-driven vs sub-test patterns, assertion
  library, and mock approach already in use.

### 2.2 When You DO Need a New Pattern

1. Point out that no existing pattern covers this case.
2. Propose the new pattern in a comment/plan before implementing.
3. Get user approval.
4. Implement and document the pattern so future agents follow it.

---

## Section 3 — Ambiguity & Business Logic

### 3.1 Never Guess Business Rules

- If the spec or user instruction is ambiguous, ask. Do not fill in gaps
  with plausible-sounding guesses.
- **"What should happen when..."** — if the spec doesn't cover an edge
  case, list the uncovered cases and ask.
- **Validation rules, authorization rules, pricing logic** — anything
  that encodes a business decision must be traceable to a requirement
  or a user confirmation.

### 3.2 Explicit Non-Goals

When designing a feature, always state what you are explicitly NOT
going to do. This prevents scope creep and clarifies boundaries.

---

## Section 4 — Output & Testing Discipline

### 4.1 Build Verification

- **After any code change, run the build.**
  ```bash
  # Go backend
  cd backend && go build ./...

  # Next.js frontend
  cd frontend && npx tsc --noEmit

  # Expo mobile
  cd mobile && npx tsc --noEmit
  ```
- If the build fails, fix it before ending your turn. Do not leave the
  user with a broken build.
- Run targeted tests for the code you changed. If you changed a handler,
  run that handler's tests. Escalate to the full suite only when
  relevant.
- **Node projects: verify Node version matches `.nvmrc`** before any
  `npm` or `node` command. The agent shell does not auto-load nvm. If the
  active Node version doesn't match, prefix commands with the nvm path:
  ```bash
  node --version  # verify before proceeding
  # If mismatch, use:
  PATH="$HOME/.nvm/versions/node/v$(cat .nvmrc)/bin:$PATH" npm install
  ```

### 4.2 Test Coverage

- New code in backend must have table-driven tests covering:
  - Happy path
  - Each error path
  - Edge case (empty input, max length, boundary values)
- New UI components must have at minimum a render test and tests for
  each distinct state (loading, empty, error, populated).
- **Test data must avoid secret-scanner patterns.** Fake keys, tokens,
  and credentials in test fixtures must not match patterns that trigger
  GitHub push protection (Stripe `sk_live_`, `rk_live_`; AWS `AKIA*`;
  GitHub `ghp_*`; etc.). Use obviously-fake prefixes like `test_`,
  `fake_`, or project-specific prefixes.

### 4.3 Linting

- Fix all lint errors before considering work done. Do not suppress
  warnings unless explicitly asked.
- **Run the linter as part of pre-push verification**, not just the
  type checker. CI enforces `--max-warnings 0`, so a clean `tsc` is
  not sufficient:
  ```bash
  # Frontend
  cd frontend && npm run lint

  # Backend
  cd backend && go vet ./...
  ```

---

## Section 5 — Git & Commit Hygiene

### 5.1 Never Commit Without Explicit Request

- Do NOT run `git commit` or `git push` unless the user explicitly
  asks you to.
- You may run `git --no-optional-locks status`, `git diff --stat`,
  and other read-only git commands freely.
- **Before committing, verify the user's message contains an explicit
  write instruction** ("commit", "push", "merge", "create a PR").
  Inspection verbs ("check", "look", "review", "show") are NOT
  authorization to commit. If unclear, ask: "Ready for me to commit
  and push this?"

### 5.2 Branching Convention

When the user asks you to start work:
- Branch prefix: `feat/`, `fix/`, `chore/`, `refactor/`, `docs/`
- Use kebab-case: `feat/user-profile-edit`
- Push the branch immediately so CI runs.

### 5.3 Commit Messages

- Follow [Conventional Commits](https://www.conventionalcommits.org/):
  `feat:`, `fix:`, `chore:`, `refactor:`, `test:`, `docs:`
- Body explains WHY, not just WHAT.
- Keep subject line under 72 characters.

---

## Section 6 — Tool Access

### 6.1 Read-Only Operations

You may freely:
- Read files with `read_file`
- Search with `grep` and `find_path`
- List directories
- Run read-only git commands
- Run build and test commands

### 6.2 Write Operations

- Edit files with `edit_file` (preferred for targeted changes)
- Create files with `write_file`
- Run terminal commands that modify the project (e.g., `go mod tidy`)
- Run `git` write operations only when explicitly asked
- **Clean up after parallel agents.** After spawning sub-agents that
  write files, check for and remove junk artifacts (files with ` 2`
  suffix, duplicated directories) before committing.

### 6.3 What You May NOT Do

- Run commands that require network access without the user's knowledge
- Install global packages without asking
- Modify configuration outside the project without asking
- Delete branches without asking

---

## Section 7 — The Specs Directory

The `specs/` directory is the **single source of truth** for all active
features. Specs are git-committed (unlike temporary working docs).

- **Read specs before implementing.** Every engineer role must read the
  relevant spec from `specs/` before writing code.
- **Specs are named**: `specs/<feature-slug>.md`
- **Spec lifecycle**: Draft → Review → Approved → Implemented → Archived
- Never modify an "Approved" spec without the PM role's sign-off.

---

## Section 8 — Session Log & Retros

Every session that involves corrections, bug fixes, or user feedback
MUST log those events to a session log file so nothing is lost before
the retro. The retro's job is to trace each correction to a specific
missing rule and fix that rule.

### 8.1 Session Log

- **File:** `specs/memories/<YYYY-MM-DD>-session-log.md`
- **Format:** a running table of corrections with root cause + fix,
  plus a checklist of questions/follow-ups
- **Update after every correction**, not just at the end
- One log per day/session

### 8.2 Retro Process

At the end of a feature or session:
1. Review the session log
2. For each correction: trace it to the specific rule that was missing
   (or didn't exist yet)
3. Update the relevant skill or AGENTS.md with the new rule
4. Create a retro file: `specs/memories/<YYYY-MM-DD>-<topic>.md`
5. Cross-link the session log to the retro

---

*Last updated: 2026-08-08*
*These rules apply to all agent threads in this project.*
