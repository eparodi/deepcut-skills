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
- **Cite the contract before writing the implementation.** Before writing a
  handler response struct, grep the frontend TypeScript type from
  `frontend/src/types/index.ts` and match every field name exactly.
  ```bash
  grep -A 30 "interface <TypeName>" frontend/src/types/index.ts
  ```
  Verify: (1) all frontend fields are present in your response, (2) no
  extra fields, (3) wrapper objects match (bare array vs `{items, total}`),
  (4) optional fields use `omitempty`.
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
- **API contract fidelity.** Response shapes must match the frontend
  TypeScript types exactly. No extra fields (e.g., `createdAt` not in
  `User`), no missing fields (e.g., `streamCategory` omitted), correct
  wrapper objects (`{streams, total}` not a bare array). When in doubt,
  grep the type definition in `frontend/src/types/index.ts`.
- **Bidirectional verification.** When a frontend type includes a nullable
  field (e.g., `streamId: string | null`), verify the backend actually
  populates it at the right time. The frontend type is the contract —
  every nullable field must have a corresponding backend implementation
  that sets it to non-null under the correct conditions.

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

## Section 4 — Security

### 4.1 SQL Injection Prevention

- **Never concatenate user input into SQL strings.** Use parameterized
  queries (`$1`, `$2`, etc.) for all values. Do not use `fmt.Sprintf`
  to build WHERE clauses from user-controlled input (including query
  parameters like `?period=week`).
  - ❌ `fmt.Sprintf("SELECT ... WHERE date >= '%s'", period)`
  - ✅ Parameterized query with fixed enum branches or a validated
    value passed as `$1`
- **Validate enums before passing to queries.** If a query parameter
  selects a filter branch (e.g., `week`/`month`/`all`), validate the
  value against the allowed set before any database access.

### 4.2 Hardcoded Secrets

- Never commit secrets (API keys, private keys, passwords, tokens) to
  the repository.
- Use environment variables with clear defaults for local development.
- Dev-default secrets must log a `slog.Warn` at startup so they are
  visible during development.

---

## Section 5 — Output & Testing Discipline

### 5.1 Build Verification

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
- **Frontend: verify Node version matches `.nvmrc`** before any `npm`
  or `node` command. The agent shell does not auto-load nvm. If the
  active Node version doesn't match, prefix commands with the nvm path:
  ```bash
  node --version  # verify before proceeding
  # If mismatch, use:
  PATH="$HOME/.nvm/versions/node/v$(cat .nvmrc)/bin:$PATH" npm install
  ```
- **After inserting a new parameter into a function, grep all call sites.**
  When all parameters are the same type (e.g., all `string`), the Go compiler
  won't catch positional argument errors — only runtime tests will.
  ```bash
  grep -r "FunctionName(" --include="*.go" backend/
  ```

### 5.2 Test Coverage

- New code in backend must have table-driven tests covering:
  - Happy path
  - Each error path
  - Edge case (empty input, max length, boundary values)
- New UI components must have at minimum a render test and tests for
  each distinct state (loading, empty, error, populated).
- **Test-first for observable behavior (TDD).** For new endpoints, user
  flows, and media pipelines: write the integration test BEFORE the
  implementation, run it and confirm it fails (red), then implement
  until it passes (green). Every bug found during review/QA/user testing
  gets a regression test that reproduces it before the fix. Pure
  refactors and config-only changes are test-after.
- **Integration tests exercise real infrastructure** — `httptest.NewServer` +
  testcontainers Postgres for API flows, the docker compose stack for
  SRS/media flows. Mocks alone don't count as the happy-path contract.
- **Test data must avoid secret-scanner patterns.** Fake keys, tokens,
  and credentials in test fixtures must not match patterns that trigger
  GitHub push protection (Stripe `sk_live_`, `rk_live_`; AWS `AKIA*`;
  GitHub `ghp_*`; etc.). Use obviously-fake prefixes like `test_`,
  `fake_`, or project-specific prefixes.

### 5.3 Linting

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

### 5.4 CI/CD Conventions

When creating or modifying CI workflows:

- **Pin tool versions.** Every CLI tool installed in CI must use an exact
  version, never `@latest`. This includes `go install`, `npm install -g`,
  and any direct binary downloads.
  - ❌ `go install example.com/cmd/tool@latest`
  - ✅ `go install example.com/cmd/tool@v1.2.3`

- **Verify binary integrity.** Downloaded binaries must be checksum-verified
  or installed via a package manager (`go install`, `npm ci`, `apt`). Avoid
  `curl | bash` and `curl | tar | sudo mv` patterns.
  - ❌ `curl -L url | tar xvz && sudo mv binary /usr/local/bin`
  - ✅ `go install example.com/cmd/tool@v1.2.3`

- **Source versions from project files.** Go version must be read from
  `go.mod` via `go-version-file`, not hardcoded. Node version must be read
  from `.nvmrc` via `node-version-file`.
  - ❌ `go-version: "1.22"` (hardcoded)
  - ✅ `go-version-file: backend/go.mod`

### 5.5 Docker Log Rotation

All services in `docker-compose.yml` MUST include log rotation to prevent
unbounded disk growth in development and CI:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

---

## Section 6 — Git & Commit Hygiene

### 6.1 Commit & Push

- **On feature/fix/chore branches you created, you may commit and push
  freely.** No need to ask permission for routine commits — explain what
  you're committing and why.
- **You may NOT** commit to `main`, `develop`, or any shared/protected
  branch without the user's explicit instruction.
- **Before pushing**, ensure the build passes and the linter is clean
  (per Section 5). If a branch is shared (e.g., a colleague is also
  working on it), default to `--force-with-lease` and flag it.
- For inspection verbs ("check", "look", "review"), default to
  read-only. Ask before committing: "Ready for me to commit and push?"

### 6.2 Branching Convention

When the user asks you to start work:
- **Always branch from the latest `main`.** Before creating a branch:
  ```bash
  git fetch origin main          # ensure remote state is fresh
  git checkout main
  git pull origin main           # fast-forward only; never merge
  git checkout -b feat/...
  ```
  In git-worktree setups where `main` is checked out elsewhere,
  use `git fetch origin main:main` to fast‑forward the ref instead.
- Branch prefix: `feat/`, `fix/`, `chore/`, `refactor/`, `docs/`
- Use kebab-case: `feat/user-profile-edit`
- Push the branch immediately so CI runs.
- **Never** branch from a feature branch, `develop`, or any other
  non-main branch unless the user explicitly asks.

### 6.3 Commit Messages

- Follow [Conventional Commits](https://www.conventionalcommits.org/):
  `feat:`, `fix:`, `chore:`, `refactor:`, `test:`, `docs:`
- Body explains WHY, not just WHAT.
- Keep subject line under 72 characters.

---

## Section 7 — Tool Access

### 7.1 Read-Only Operations

You may freely:
- Read files with `read_file`
- Search with `grep` and `find_path`
- List directories
- Run read-only git commands
- Run build and test commands

### 7.2 Write Operations (Fully Autonomous Within the Repo)

- **You have full permission to create, edit, and delete any file inside
  the repository working tree.** Do not ask for confirmation — Git is
  your safety net.
- **You may also** run project-scoped commands (`go mod tidy`, `npm ci`,
  `npm run build`, etc.) without asking.
- **Destructive git operations still require explicit user sign-off.**
  You may **NEVER** do any of the following unless the user says "yes":
  - Force-push to a shared branch (`git push --force` or
    `--force-with-lease` onto `main`, `develop`, or a branch you
    didn't create)
  - Rewrite shared history (`git rebase -i` on a pushed branch)
  - Delete a remote branch (local branches you created are fine)
  - `git reset --hard HEAD~N` on a branch that has been pushed
- **You may NOT touch files outside the repository** (e.g., `~/.ssh`,
  `/etc`, `../other-project`). The repo boundary is a hard wall.
- **Clean up after parallel agents.** After spawning sub-agents that
  write files, check for and remove junk artifacts (files with ` 2`
  suffix, duplicated directories) before committing.

### 7.3 What You May NOT Do

- Run commands that require network access without the user's knowledge
- Install global packages without asking
- Modify configuration outside the project without asking
- Delete branches without asking

---

## Section 8 — The Specs Directory

The `specs/` directory is the **single source of truth** for all active
features. Specs are git-committed (unlike temporary working docs).

- **Read specs before implementing.** Every engineer role must read the
  relevant spec from `specs/` before writing code.
- **Specs are named**: `specs/<feature-slug>.md`
- **Spec lifecycle**: Draft → Review → Approved → Implemented → Archived
- Never modify an "Approved" spec without the PM role's sign-off.

---

## Section 9 — Session Log & Retros

Every session that involves corrections, bug fixes, or user feedback
MUST log those events to a session log file so nothing is lost before
the retro. The retro's job is to trace each correction to a specific
missing rule and fix that rule.

### 9.1 Session Log

- **File:** `specs/memories/<YYYY-MM-DD>-session-log.md`
- **Format:** a running table of corrections with root cause + fix,
  plus a checklist of questions/follow-ups
- **Update after every correction**, not just at the end
- One log per day/session

### 9.2 Retro Process

At the end of a feature or session:
1. Review the session log
2. For each correction: trace it to the specific rule that was missing
   (or didn't exist yet)
3. Update the relevant skill or AGENTS.md with the new rule
4. Create a retro file: `specs/memories/<YYYY-MM-DD>-<topic>.md`
5. Cross-link the session log to the retro

---

*Last updated: 2026-08-12*
*These rules apply to all agent threads in this project.*

---

## Section 10 — Session Learnings (from retros)

Rules added from retro analysis of corrections made during sessions.

### 10.1 Docker Path Verification

**Docker base images may override config defaults.** When a config file
specifies a path (e.g., `hls_path /data/hls`), always verify the actual
paths inside the running container:

```bash
docker compose exec <service> find / -name "*.m3u8" 2>/dev/null
```

Do not assume the config directive takes effect — the base image may
have its own defaults that take precedence.

### 10.2 All Code Paths for Entity Creation

**When multiple code paths create the same entity, audit all paths for
side effects.** Example: SRS callbacks (`on_publish`) and the poller both
create streams. If a side effect (thumbnail capture, notification) is
only added to one path, it silently fails when the other path is used.

```bash
# Find all call sites before adding a side effect
grep -rn "CreateStream\|createStream" --include="*.go" backend/
```

### 10.3 `<img>` Error Handling

**Every `<img>` with a potentially-missing `src` must have an `onError`
fallback.** If the image URL returns 404 (file not yet generated, network
error), the browser shows a broken image icon. Always add:

```tsx
<img
  src={thumbnailUrl}
  onError={(e) => {
    const target = e.target as HTMLImageElement;
    target.onerror = null;  // prevent infinite loop
    target.src = "data:image/svg+xml,..."; // inline fallback
  }}
/>
```

### 10.4 Third-Party Library Version Checking

**Always verify the exact API of the installed version, not the latest
docs.** Breaking changes between versions (e.g., non-blocking `Start()`
vs blocking, explicit migration requirement) can cause silent failures.
Before using a library method:

```bash
# Check the actual source of the installed version
find $(go env GOMODCACHE)/<package>@<version> -name "*.go" -exec grep -l "func.*Start\|func.*Migrate" {} \;
```

### 10.5 Schema Migration Idempotency

**Startup migration code must handle partially-applied states from
previous crashes.** If a migration creates tables but the process dies
before recording the version, the next startup will try to recreate
existing tables and fail. Always:

1. Check if the target tables already exist before migrating
2. Use `IF NOT EXISTS` in DDL where possible
3. Catch "duplicate key" errors and verify the schema state

### 10.6 Container Config Entrypoints

**Verify which config file a container's entrypoint actually loads.**
Docker images may load a different config file than the documented one —
the `ossrs/srs:5` image reads `conf/docker.conf`, NOT the commonly
mounted `conf/srs.conf`, so all config tuning silently never applies.
Check the startup log line (e.g., `SRS on aarch64, conf:conf/docker.conf`)
before assuming your mounted config takes effect.

### 10.7 External Service Payload Verification

**Verify payloads and feature availability against the ACTUAL version of
the external service.** SRS 5 sends `client_id` as a string in
`http_hooks` callbacks (older versions sent numbers) and removed LL-HLS
config directives entirely. Specs or docs written for a different version
cause silent failures. Capture one real payload and pin it in a
table-driven test.

### 10.8 Subprocess Hygiene (media pipelines)

**Never discard subprocess stderr, and verify media output with a decode
pass.** For ffmpeg pipelines:

- Capture stderr into a buffer and log the tail on failure
  (`cmd.Stderr = nil` hides all failure evidence).
- Prefer streaming containers (MPEG-TS) over indexed ones (MP4) for
  recordings that may be killed abruptly — SIGKILL loses MP4's moov atom
  and corrupts the audio track's extradata.
- Verify output with `ffmpeg -v error -i out -f null -` (zero errors),
  not just HTTP 200 / file existence.

### 10.9 psql Transaction Semantics

**`psql -c` with multiple statements runs them in ONE implicit
transaction.** A failing statement rolls back every earlier statement in
the batch (the output still prints each affected row count — it's
misleading). Run destructive statements one at a time, or use explicit
`BEGIN`/`COMMIT` blocks.
