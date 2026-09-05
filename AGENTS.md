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
  `User`), no missing fields (e.g., a required category field omitted),
  correct wrapper objects (`{items, total}` not a bare array). When in doubt,
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
  media/streaming flows. Mocks alone don't count as the happy-path contract.
- **Integration tests sharing one database must not run packages in
  parallel.** `go test` runs packages concurrently by default; when every
  package truncates the same tables, one package wipes another's rows
  mid-test (flaky FK violations / missing rows). Run with `-p 1` in CI,
  or give each package its own schema/database.
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

*Last updated: 2026-09-05*
*These rules apply to all agent threads in this project.*

---

## Section 10 — Shared Session Learnings (from retros)

Generic rules from retro analysis of corrections made across the
repos. They apply to EVERY repo (code style, tool discipline,
UI, ops, payload verification, deploy ordering) — project-specific
learnings stay in each repo's own AGENTS.md. Cite them as
"skills-test AGENTS.md §10: <rule name>".

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
side effects.** Example: a media server's publish callback
(`on_publish`) and a poller both create stream records. If a side
effect (thumbnail capture, notification) is only added to one path, it
silently fails when the other path is used.
Same trap in a trading system: trades open from live execution, dry-run,
AND startup reconciliation — if a side effect (learning, stats update,
exit-rule capture) is added to only one path, it silently fails on the
others.

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
docs.** Breaking changes between versions can cause silent failures:
non-blocking vs blocking `Start()`, an explicit migration requirement,
`Tx.NextSequence` vs `Bucket.NextSequence` (bbolt), API drift in
`math/rand/v2`. Before using a library method:

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
config directives entirely; Binance klines and DeepSeek chat responses
also drift between API versions and model tiers (a model emitting a
chain-of-thought into `reasoning_content` changes the max_tokens math).
Specs or docs written for a different version cause silent failures.
Capture one real payload and pin it in a table-driven test before
trusting docs.

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

### 10.10 Config Deploy Ordering

**Deploy the binary/service before adding new config keys.** Config
parsers that reject unknown fields (Go `json.Decoder` +
`DisallowUnknownFields`; strict pydantic/serde) crash or crash-loop a
service when the live config contains a key the deployed binary doesn't
define yet. Order for any config-key change: (1) deploy the new
service, (2) verify it starts cleanly, (3) add the key to the live
config, (4) restart. Reverse order = outage (a supervisor with
restart-always turns it into a crash loop). Same rule bit a remote
deploy: `deploy.sh` ships the binary first and leaves config.json
untouched — verify the service actually started, then add the field
and restart.

### 10.11 Claimed Edits Must Be Verified

**Never report an edit as applied without the tool result proving
it.** A previous session stated two spec edits were "updated" when no
edit tool had run — the file was unchanged and the correction was only
caught by re-reading. After claiming an edit, re-read the affected
section or trust ONLY the tool's success result, never memory.

### 10.12 SKILL.md Frontmatter Must Be YAML-Parsed, Not Eyeballed

**Unquoted `description:` values containing `: ` (colon+space) are
invalid YAML** — `example: the provider...` breaks the
frontmatter at line 3 (`mapping values are not allowed`), and the
skill fails to load. After writing any `SKILL.md`:

```bash
ruby -ryaml -e 'ARGV.each { |f| YAML.load_file(f); puts "#{f}: OK" }' .agents/skills/*/SKILL.md
```

Quote any description containing `: ` with double quotes, then re-parse.
Do not trust a visual check — a YAML parser is the only verification.

### 10.13 Append-At-End Edit Anchoring

**Appending code to an existing file must anchor on the file's actual
LAST unique line (read the tail first).** Generic closing-brace
patterns (`}\n\t})\n}`) match many functions — the edit tool inserts
mid-file and breaks brace structure. Symmetric old/new anchors can
silently delete the anchored lines; after every multi-edit batch,
re-read the affected regions (matches 10.11).

### 10.14 Comment-Only Contracts

**An input-order contract that lives only in a doc comment WILL be
violated by a caller or a test.** When ordering matters (newest-first
input, deterministic output), enforce it INSIDE the function (sort,
stable) — robustness belongs in the function, not in the comment.
Also: fixtures writing to age-pruned stores must use `time.Now()`-based
timestamps — epoch-1970 rows are correctly pruned and fail tests
mysteriously.

### 10.15 Float Display Formatting

**Trimming trailing zeros after `strconv.FormatFloat` must only touch
fractional digits.** `strings.TrimRight(s, "0")` on a value formatted
with `prec=0` (no decimal point) eats integer zeros: 100 → "1", 1e8 →
"1M". Guard the trim with `strings.Contains(s, ".")`, or always format
with `prec > 0` so a decimal point exists. Pin regression rows for
round powers (100, 1e8) in the formatter's table tests.

### 10.16 Pinned Copy and Ban Strings in Page Tests

**Unescape `html/template` entities before comparing pinned copy** in
rendered HTML (`bot&#39;s` ≠ `bot's` — still a recurring trap). **Ban
test strings must target rendered VALUES, never labels or links** —
a term that legitimately appears as a glossary link is not the thing
the ban asserts; the value is what must be absent. **Spec-pinned copy
renders verbatim on every profile/variant** unless the spec scopes it
(a /backtest empty-state run instruction).

### 10.17 Test Configs Must Survive the Validation the Code Runs

**When the code under test validates a candidate config, the test-env
config must be validation-clean** — invalid floors or a zero config
block break tests far from their cause. Keep secondary test-only
weighting in a SEPARATE map so the validated config stays valid
without changing the behavior under test.

### 10.18 Shared Lists Need Per-Consumer Pinning

**When one list feeds two different render consumers, filter it per
consumer — never pass a zero-value struct into a template define.** A
zero-value struct passed to a define renders REAL, visible DOM (empty
label + empty input), not nothing. Pin each consumer's rendered output
with a render test, including what must be ABSENT (a phantom empty
`num-input` under `<label for="symbols">` is the cautionary tale
— `riskFields` fed both the confirm page's rows and the settings form's
order; the form never filtered `symbols`).

### 10.19 Behavior-Neutral Refactors Ship with a Rendered-Output Diff

**When a refactor claims to change nothing visually, prove it with a
rendered-output diff.** Render the pre-change commit (worktree + a
temporary dump test) and the change into files and diff them — only
intended deltas may appear (CSRF tokens are the expected noise). Clean
up the dump test and the worktree afterwards. An eyeball check is not
evidence; the diff immediately exonerates or convicts the refactor.

### 10.20 Model Tier Routing

**Mechanical, reversible tasks run on Flash; judgment or
hard-to-reverse tasks run on Pro.** Mechanical = applying spec'd
edits, running builds/tests, grep/summarize, session-log and retro
drafts, boilerplate. Pro = architecture and API contracts, security
audits, money/risk rules, debugging after Flash exhausts its repair
ladder, skill design (factory), learning distillation (porter). Flash
escalates after 1 local repair + 1 re-ask, handing over findings so
Pro never re-reads context Flash already read.

### 10.21 Pre-Existing Claims Are Proven Against the Merge-Base

**"Fails without my changes" is only proven by testing the merge-base
commit — never a stash of uncommitted work.** A stash check leaves
EARLIER COMMITTED changes (from the same workstream) active, so the
failure still contains your bug and gets misattributed as
pre-existing (2026-08-15: the strict pricing validation was already
committed when the stash test "proved" two failures pre-existing —
they were mine). Use `git stash` only for uncommitted changes; for
committed ones, test `git merge-base main HEAD` or the parent commit.

### 10.22 Insert-Before Edit Anchoring

**When inserting new code before an existing block, never anchor on
the block's first line alone.** Replacing `func X(...) {` with the new
code + `func X(...) {` silently deletes X's opening body lines (three
recurrences 2026-08-15, each repaired by re-reading). The safe
pattern: old_text = the block's signature PLUS its first body lines;
new_text = new code + that same prefix verbatim. Same family as 10.13
(append-at-end) — re-read the affected region after the edit either
way (10.11).

### 10.23 Partial Edit Batches Must Be Reconciled From the Diff

**A multi-hunk replace that partially lands leaves the file with MIXED
old/new signatures — continuing without re-reading breaks the build
silently (2026-08-18: a file kept its old function signatures after a
batch whose last hunk failed to match; only one section had been
rewritten).**
After any edit batch where a hunk failed or was reworked, read the
file (or `git diff`) and reconcile EVERY caller of the changed
signature before building. Same family as 10.11/10.13/10.22 — the
compile error is the cheap symptom; the expensive one is a
partially-migrated file that still compiles.

### 10.24 Test Configs Run the Production Config Pipeline

**A config layer that synthesizes defaults or validates during a
`Validate()`/`Load()` step MUST be exercised by test fixtures the same
way production does.** Building directly from hand-constructed structs
bypasses defaults synthesis (2026-08-18: tests failed a required-agent
validation — the registry defaults are synthesized in
`Validate`, which the tests skipped). `testCfg` helpers must call the
same entry point as `main`; assert the synthesized defaults in the
test, not by accident.

### 10.25 Stateful Fakes Advance Time Per Request

**Fakes that feed time-sequenced logic must advance their
time-dependent state on every request** (2026-08-18: a fake exchange
served the SAME 62 klines every call — the snapshot TS never
advanced, so a journal's `signalTS < ts` scoring could never
fire; 2023-era timestamps were additionally age-pruned by the store's
30-day prune the moment they were saved). Candle windows, pagination
cursors, and "next" tokens in fakes must move forward like the real
service; anchors belong at `time.Now()`-ish values (extends §10.14).

### 10.26 Budget-Gate Tests Model the Exhaust→Veto Sequence

**Budget/circuit gates check `Exceeded()` at the START of the next
call — the call that exhausts the budget PASSES, the veto lands on the
next call** (2026-08-18: the budget-fired-day test needed cycle 1 to
exhaust + cycle 2 to veto). Also: test builders must wire the same
callbacks production wires (token metrics → the budget adder); a budget
that never receives token counts never vetoes. Sequence tests without
the exhaust step assert a gate that never fires.

### 10.27 Probes Target Markup, Never Bare Class Names

**When component stylesheets are bundled and inlined into every
page's `<head>`, a class-name probe matches the CSS bundle regardless
of the rendered markup — position assertions (e.g., "banner topmost")
become meaningless** (2026-08-18: a banner test matched a component
class name inside the stylesheet). Assert on markup
(`<div class="scoreboard-grid">`, `<table …>`), and unescape
`html/template` entities before comparing pinned copy (§10.16 — `+`
renders `&#43;`).

### 10.28 bbolt Bucket Sweeps Enumerate Dynamic Buckets and Empty, Not Delete

**A static bucket-name list is blind to per-key namespaced buckets.**
`<topic>:<agent>:<symbol>` / `<topic>_cursor:*` buckets
are invisible to a literal `"<topic>_cursor"` entry — the
"wipe all tracking data" sweep matched nothing for them, and a
surviving cursor would make a re-distiller reprocess stale rows
(2026-08-18, a `data clean` command). Sweep with `tx.ForEach` + prefix
matching. Two more bbolt traps in the same family:

- **Empty buckets, never delete them** — writers that use plain
  `tx.Bucket(name).Put` (no `CreateBucketIfNotExists`, e.g.
  `SaveEquityPoint`) nil-panic on a deleted bucket.
- **Read-only opens only serve pure `View` reads** — readers that
  lazily create buckets via `Update` + `CreateBucketIfNotExists`
  (e.g. `Learnings`) fail with "database is in read-only mode";
  verify the read path is View-only before using a read-only handle.

### 10.29 Shared State Blobs Are Patched, Not Replaced; Backups Never Overwritten

**A full-replace save of a shared aggregate silently drops the fields
other subsystems persist.** A shared `state` bucket carries the
breaker fields AND the LLM daily-budget rollover, fired-alert days,
and the summary day (2026-08-18: a reconcile CLI saved
the raw breaker state, wiping the persisted `$2/day` LLM cost cap —
the next restart restored an empty budget). One-off CLIs/tools persist
what they own: **load → patch → save**. Same family: a destructive
tool with a timestamped backup must refuse to overwrite an existing
one — a same-day second run destroys the only pre-destruction snapshot
(same filename collision).

### 10.30 Claimed State Transitions Need a Test That Exercises Them

**If a comment claims a flag clears on success, write the test that
proves it — a transition you cannot test is a transition that cannot
happen** (2026-08-18: a stale flag's clearing function claimed a
successful close clears it, but every exit path gates on
a block flag BEFORE the close, so it was unreachable dead
code; the approved semantics were sticky-until-restart). Either
remove the dead code and pin the ACTUAL semantics with a test
(restored balance → still no retries → flag persists until
restart/reconcile), or restructure so the test passes for the right
reason.

### 10.31 At the Approval Gate, Restate the Operator's Original Vision

**The operator's original intent can drift out of the reviewed
artifact** (2026-08-18: a seconds-level real-time reaction vision
was recognized as a mismatch only AFTER a design that dropped it
shipped — the gate had decided a derivative). At the approval gate,
restate the operator's core ask as explicit acceptance criteria and
record consciously-deferred variants (e.g., "real-time mode") as
explicit non-goals or follow-ups — the gate must decide the ORIGINAL
ask, not a derivative the discussion drifted to.

### 10.32 Records of Exchange Balances Must Be Net-Of-Fee

**Spot exchanges charge trading fees in the RECEIVED asset** — a
BUY's fee comes out of the base asset, so the exchange holds
`ExecutedQty − base commission`, not the gross fill (2026-08-20
live-mainnet: BNB recorded 0.041000 vs held 0.04079267, 0.5057% —
full-qty SELLs failed `-2010` and a tolerance-less phantom check
marked the position stale, stopping exits for 14h). Two-sided rule:
(1) persist the NET quantity derived from per-fill commissions
(`commissionAsset` == base asset), and (2) keep the fee in PnL
(fees = quote commissions + baseFee × fill price) — netting the qty
without adding the fee back overstates PnL by the fee principal.
Verify the fee currency from the fill payload, never assume it.

### 10.33 One-Record Caches Undercount Multi-Record Stores

**An in-memory view keyed by an entity that holds ONE record while
the store keeps every record silently undercounts every aggregate**
(2026-08-20: a cache keyed by symbol kept the last buy; equity,
exposure caps, and dashboards saw 24.93 of 131.4 USDT — the 30%
cap was defeated and older records lost TP/SL protection until a
restart reloaded them one at a time). When a store is per-record,
any cache/view of it must aggregate (slices, sums) — and exits,
valuations, and published state must consume the aggregate, not a
sample. The DB being right while the view is wrong is still a bug
that loses money protection.

### 10.34 Claimed Facts Trace to Code or Evidence, Never Recollection

**Two readings in one session:** (1) a claim that "trading started
yesterday" was off by 1.5 days — the real-money switch was proven
from `.env` mtimes and config backups (2026-08-18 14:55, not 08-19);
(2) a "deposit mystery" (day equity 94.46 vs total 174.56) resolved
when the field was read in code — a day-snapshot field was MANAGED
equity (free USDT + tracked positions), not the total. Interpret
snapshot/dashboard fields by their code definition, and date
"when did X start" claims from file mtimes, config backups, or
deploy logs. A field name is a hypothesis, not a fact.

### 10.35 Fixture Formatting Precision Is Part of Pinned-Test Contracts

**Pinned tests can depend on a test double's exact output rounding**
(2026-08-20: the net-zero SL test computed its fill price from the
fake's `%.4f` commission rounding; changing the fake to `%.8f` broke
the exact-zero assertion). Changing fixture formatting is a contract
change: check every pinned test that feeds on the formatted value,
or make the precision explicit per field (e.g., 4 decimals for USDT
commissions, 8 for base-asset fees) so the rounding intent is
visible.

### 10.36 Test-Env Seed Persistence Gates Silently Drop Fixtures

**A test harness that persists its seed state only when some sentinel
field is non-zero silently drops a fixture whose new/zero-valued
fields don't trip the gate — the test then passes against an unseeded
store** (2026-08-22: a fixture seeded with only a new state field
never reached the store until a sentinel was set). A fixture that must
be restored must also satisfy the gate's condition, or the gate must
be widened for the new field type.

### 10.37 Rollover-Driven Accumulators Must Survive the Process's First Save

**The first save of a process lifetime can run before a rollover
driver has produced its first value — an unconditional write of the
in-memory accumulator then clobbers the restored value** (2026-08-19:
the startup save reset a restored LLM lifetime bank before the first
budget roll). Tracking fields must update only when the driver yields
a non-empty value.

### 10.38 Verify the Worktree HEAD Before Branching

**`git fetch origin main:main` moves the ref, not a detached
worktree's HEAD — `git checkout -b` in that worktree silently bases
the branch on an older commit** (2026-08-19: a feature branch was cut
3 commits behind main, forcing a mid-task rebase). Compare
`git rev-parse HEAD` with the target ref before branching; re-checkout
if they differ. **The same HEAD check applies before any build/deploy
step that consumes the worktree's checkout** — a deploy script that
cross-compiles from `$PWD` ships whatever ref the worktree sits on
(2026-08-24: a deploy shipped a pre-merge binary), and a green
startup log proves the OLD binary can start, not that the new code
shipped; verify the shipped artifact with a feature marker
(`strings <binary> | grep <marker>`), never "it started" alone.

### 10.39 Multi-Series Charts Need a Legend and a Per-Series Palette

**A chart that draws more than one series must render a legend and
assign each series its own color** — otherwise every series draws in
one color and end-of-line labels collide when lines finish near the
same point (2026-08-20: a multi-strategy chart rendered every
strategy in the same blue).

### 10.40 Bar Width Must Be Gap-Aware

**Bar width derived from the series count (not the x-gap) draws
clustered bars on top of each other and makes lone bars comically
fat** — width = min(count rule, 0.8 × smallest adjacent gap), sorted
by X inside the renderer (2026-08-20: clustered bar charts; fixed with the
gap-aware rule).

### 10.41 Themed SVG: Split Class Names by CSS Property

**When theming inline SVG via CSS classes, class the data colors by
CSS property (`stroke-c-*` vs `fill-c-*`) so a stroke rule can never
clobber a path's `fill="none"`** (2026-08-20: chart dark-mode work —
the property split prevents a blanket color rule from filling unfilled
paths).

### 10.42 Bundled Enhancements Need Per-Page Presence Pins

**A controller shipped in a shared JS bundle does nothing on any page
that fails to load the bundle — "shipped" features are dead everywhere
except the page where the script tag was first added** (2026-08-20:
chart tooltips never ran on 7 of 8 pages because the tag lived only in
one template). Load the bundle in the SHARED layout and pin its
presence on every page.

### 10.43 Hover/Tooltip Data Lives on the Visible Element

**Tooltip data (`<title>`) attached to an invisible helper shape is
dead UI — the user never points at it** (2026-08-20: a 4px invisible
hover circle carried the bar tooltips). Put the data INSIDE the
element the user actually hovers.

### 10.44 Backfill Paths Must Preserve the Entity's Temporal Position

**When a record carries an explicit temporal position (a date), that
position IS the timestamp — the recording moment is only the fallback**
(2026-08-20: three dated capital flows were stamped `now` and rendered
as a vertical line). Backfill/import paths must stamp from the
entity's date, and corrupted live data is repaired in place (backup
first).

### 10.45 Text-Only Models Need a File-Based Vision Pipeline

**Embedding a screenshot PNG into a text-only model's context crashes
the next call (`UNSUPPORTED_CONTENT`)** — visual QA must render to a
FILE and pass plain-text references to it (2026-08-21:
`browser_screenshot` crashed a QA session twice; the file-based
pipeline recovered it).

### 10.46 Test Doubles Must Read Configurable Flags Under the Lock and Fail Loudly on Malformed Fixture Values

**A behavior-switching flag read outside the mutex that guards the
state it controls is a data race that only fires under `-race`; a
malformed fixture input parsed silently (e.g. `ParseFloat` → 0) turns
a broken fixture into a quietly-passing test.** Capture the flags under
the lock, and make malformed fixture values fail the request loudly
instead of degrading to a default (2026-08-24: a test fake read its
flags after releasing the mutex, and a malformed `sellExecQty` parsed
as 0 silently).

### 10.47 Test Doubles That Store Derived State in Memory Can Mask a Real-Store Persistence Gap

**A fake that keeps a derived identifier (seq, cursor, id) in memory
can pass restart/reopen tests while the real store never persists
it — the recovery path then breaks only in production.** Any seam
method that returns a derived value needs a reopen→restore→use pin on
the REAL store (close, reopen, assert the value survived and the
follow-up operation works), never just the fake (2026-08-24: an
append marshalled a record before stamping its id, so every persisted
record carried id 0 while the in-memory fake made the restart test
pass).

### 10.48 The Shared Skills Repo Must Be Free of External Project and Org References

**The shared skills/learnings repo is standalone: its files must
contain zero references to external projects (by name or via their
internal identifiers) and zero org names** — skill files, the agent
index, shared §10 rules, research docs, and even historical specs are
all in scope (2026-08-25: the hub repo was scrubbed of every project
name, org name, project file path, and project-only example —
"Reference implementation" lines, a PR-review attribution, the
index's Where column, and learnings citing "the bot's fake exchange"
were all generalized). Distilled rule bodies must generalize examples
so they only make sense in the source project; project/org names
belong in the citation line only. The learning-porter's
Classify/Generalize rubric enforces this at distillation time; this
rule keeps the rest of the repo honest.

### 10.49 Process Roles Don't Carry Stack-Specific Code Details

**Process/meta roles must not carry stack-specific code details —
those belong in the stack skills** (2026-08-25: the orchestrator's
Phase 0 carried a Node-pin step with an nvm `PATH` one-liner that the
Next.js/Expo stack skills already own; the same audit found
qa/reviewer/security skills carrying hardcoded go/npm/docker commands
and project ports). A process role discovers what the repo declares
and loads the matching stack skill for the how; the one-liner lives
exactly once, in the skill for that stack.

### 10.50 Redirect Assertions Need the Unfollowed Status and Location

**When a test asserts a redirect, the client's default behavior may
follow it silently — "200" and "302" claims can both be wrong if the
harness auto-follows** (2026-08-26: a `urllib`-based probe reported
200 for every no-cookie request because the 302 → /login had already
completed). Assert with a no-redirect transport/opener and check
**both** the status (301/302) **and** the `Location` header; never
infer a redirect from the final status alone.

### 10.51 Assert Class Attributes by Full Value or Token, Never a `class=`-Prefix Substring

**Asserting `class="card"` fails to match a rendered multi-class
attribute `class="panel card"` even when the element is present — and
can pass by accident when the target class is listed first**
(2026-08-26: a probe for `class="prefs"` missed the rendered
`class="panel prefs"`, misreporting a rendering bug that didn't
exist). Assert the full attribute value, or match the token with
boundaries (`class="[^"]*\btoken\b[^"]*"`); never rely on a
`class=`-prefix substring.

### 10.52 Behavior Changes Must Update the Tests Pinning the Superseded Contract

**A feature that deliberately changes observable semantics fails
every test that pinned the OLD behavior — a green suite before the
change proves nothing about pin currency** (2026-08-26: a logout test
still asserted server-side session revocation after the redesign made
logout a cookie clear). When the change lands, grep the superseded
contract's terms in tests (old sentinel values, status codes, response
fields) and update those pins in the same change; then pin the new
semantics explicitly.

### 10.53 Time-Pinned Fixtures Need Now-Relative Anchors

**A fixture that pins a computed output (a delta, a bucket, a day
label) must anchor its time-dependent inputs at now-relative values —
a fixed wall-clock anchor is in the future for part of every day, so
the input falls outside a now-based range and the pinned output
drifts silently across the day boundary** (2026-08-26: a "today at
12:00 UTC" point dropped out of [today−30d, now] before noon,
drifting a pinned +2.0% delta to +1.0%; extends the
§10.14/§10.25 time-anchor family). Anchor with `now − offset` values
that are always inside the range, or run the time-sensitive test on
both sides of midnight UTC.

### 10.54 Markdown Structure Parsing Must Be Fence-Aware

**Extracting structure from markdown (headings, links, outlines) must skip
fenced code blocks — template examples inside ``` fences are content, not
structure** (2026-08-26: a generated catalog listed a skill's
template-example headings as six extra sections). A naive extractor that
matches heading syntax anywhere in the file silently lists placeholder
sections from embedded templates as if they were real. Pin the real
section counts of template-embedding documents in a test.

### 10.55 Docs-Deliverable Form Factor Is a User Decision, Not an Assumption

**Before designing a wiki/docs deliverable, confirm the delivery surface
(in-repo docs, GitHub wiki, hosted site) with the user — never bake a
form-factor non-goal into a spec without asking** (2026-08-26: a wiki
spec's "no GitHub wiki" non-goal contradicted the operator's actual
intent). The requirements gate is the cheapest place to learn the surface;
the shipped artifact is the most expensive place.

### 10.56 Time-Window Boundaries Derive From the Enforced Grid and Carry a Full Step of Latency Margin

**A "next boundary" window can be 1 second away from the entity it
must exclude — not a full interval** (2026-08-27: a trading bot's
`ExitCycle` is the closed candle's LAST second (`CloseTime/1000 =
open+899`), so a window opening at the next boundary still contained
the close's own fills and double-booked the ledger when a manual sell
also existed). Rules: (1) derive the step from a VALIDATED constant
(the candle grid pinned by config validation), never a free config
knob; (2) open the window a full step past the entity's interval so
processing/fill latency is covered; (3) re-derive timestamp semantics
from code + real payload shape before trusting a review's boundary
assumptions.

### 10.57 Fakes for Server-Side-Filtered Endpoints Must Honor the Filter Params

**A fake that returns every record regardless of the query's
`startTime`/`endTime` silently bypasses the very windowing the code
under test implements** (2026-08-27: a reconcile fake served all
trades with a stub timestamp, so the double-booking window was never
exercised). If the real service filters server-side, the fake must
parse the filter params and apply them inclusively — otherwise the
test proves nothing about the window (extends §10.25, which is about
fakes ADVANCING their own time state; this is about honoring the
client's requested window).

### 10.58 Prefer Exclusion-by-Construction Over a Stateful Skip-Budget When Reconciling a Stream Against a Ledger

**A per-consumer qty budget for skipping already-recorded stream
items double-books when fills interleave** (2026-08-27: a reconcile's
skipQty budget reset per record and a positional head-skip booked the
bot's own TP fill twice). Filtering the SOURCE (open the stream past
the last recorded item) is stronger than maintaining shared skip
state across consumers — state resets, positionals mis-book, and the
same item can be allocated twice.

### 10.59 A 200 With an Unparseable Body Is Transient — Retry Within the Budget

**A status-code-only retry ladder misses connection-cut bodies**: a 200
whose JSON was truncated mid-write is a transient provider failure, not
an application error (2026-08-27: the visual-QA vision client hard-
failed two runs with `unexpected end of JSON input` until unparseable
200s became retryable, pinned by a retry-and-exhausted test pair). Treat
HTTP 200 + failed decode as retryable through the same budget as
429/500/503; only error out once the budget is exhausted.

### 10.60 Pin Provider Budgets Against Live Measurements and Vendor Docs — Never Assumptions

**Reasonable-sounding budgets can be 10× off from reality**: a 30s
client timeout vs 26–33s real vision responses, `max_tokens: 4096`
exhausted by 2k–4k reasoning tokens alone (empty content), and an
assumed 16384px Chrome capture cap when the binding limit was the
vision API's documented 8192 px/side (400 "unsupported image") — all
pinned wrong before the first full-checklist live call (2026-08-27,
visual-QA). Probe the live provider (measure latency, token burn,
limits) and read the vendor limits page before fixing timeouts, token
caps, or image-size guards — then pin the measured numbers in tests.

### 10.61 Content Injected Into a Prompt Is Untrusted — Say It's Data

**Any content channel that injects external text into an LLM prompt
(page HTML, fetched documents, user-generated strings) is a prompt-
injection surface** — the system prompt must explicitly say the
injected content is data to be evaluated, never instructions (2026-08-27:
the visual-QA HTML-evidence channel; the bot repo's §4.2 learnings rule
is the same discipline in one repo). Add the sentence when the channel
ships, not after an incident.

### 10.62 json.Decoder.Decode Ignores Trailing Content — Require EOF for Strict Parsing

**`json.Decoder.Decode` reads ONE value and silently ignores the rest**
— a strict parser must decode a second time and require `io.EOF`, or
`{...} extra` passes validation (2026-08-28: the visual-QA explorer's
`next_action` parser, and the same latent gap in `parseChecks`).

### 10.63 Driver Loops Must Reach Their Start State Before the First Observation

**A loop that observes or grades page state must navigate to its entry
point BEFORE the first iteration** — a driver loop that captures
before navigating grades `about:blank` and "completes" instantly
(2026-08-28: the visual-QA explore loop; the model saw an empty page
and replied `done` until the start-URL navigation ran first).

### 10.59 Remote/Background Runs: Detach, Self-Safe Kills, and File-Based Payloads

**Three remote-ops traps from one 2026-09-02 measurement session** (an A/B
corpus run as offline jobs on a remote deploy host, live service untouched):

1. **Long jobs must run detached — never hold the ssh session open on the
   job.** Launch with `setsid <cmd> >out 2>&1 < /dev/null` and RETURN; poll
   with separate short calls. A batch controller tied to the ssh session
   dies with the pipe (or the client timeout), orphaning only its children
   and silently stopping the batch after the first run.
2. **`pkill -f` matches your OWN command line** when it contains the search
   text — the cleanup command kills the very session running it. Use a
   bracket pattern (`pkill -f 'backtes[t]'`) so the pattern text does not
   match itself, or capture and kill explicit PIDs.
3. **Multi-layer quoting (local shell → ssh → remote shell → heredoc /
   `--body` strings) mangles escapes** in one direction or another. Write
   scripts and long payloads to files and `scp` them (or use a CLI's
   `--body-file`) instead of inline heredocs; verify the remote file's
   execution separately from its transfer.
4. **Offline/measurement config toggles never live in the live service's
   config** — a feature flag flipped there becomes live behavior on the
   next restart, mid-measurement. Run offline tools from a scratch
   directory holding a copy of the config (many loaders read the config
   relative to the process CWD, so running from the scratch dir is the
   whole trick), and validate the scratch config with one tiny run before
   any spend.

### 10.64 Money-Total Displays Enumerate the System's Actual Buckets — a Plausible Mechanism in a Comment Is Not a Model

**A money-total display must sum the system's ACTUAL money buckets as the
owning code defines them — where each bucket's balance lives (the spot
wallet vs a separate-product balance API), how money moves between buckets
— and label each component** (2026-09-05: a "total on account" row
promised "plus parked when any" but modeled the parked money as a claim
token already in the spot balances, when the owning code's model parks
into a SEPARATE product fetched by its own API — the total silently
omitted the largest bucket on the live account while sitting one row above
an equity figure that includes it). Before summing money for display,
trace every bucket to the owning code (the feature spec, the client, the
valuation path) — never infer the model from a field name or a comment —
label each component so the row reconciles against the figures beside it,
and pin the composition with the discriminating fixture. The undercount-
view family in money form: a comment that names the wrong mechanism hides
real money.

### 10.65 Spec-Promised Output Is Pinned on Every Path — the Degraded One Included

**When a spec or amendment promises a row/message on every code path, EACH
path — the degraded/error one included — must render the promised content,
pinned against the PROMISE, not against what the code happens to do**
(2026-09-05: a ledger-unreadable early return dropped the spend rows its
own amendment promised on exactly that path; the pin asserted only the
note + the one surviving row, so the drop shipped green). Write the pin
from the spec's promise first ("rows X and Y still render when the ledger
is unreadable"), watch it fail red on the early-return path, then make the
code pass — a test that asserts the code's actual output instead of the
spec's promise cannot catch the promised-row drop.

### 10.66 One Entity's Rows Get One Shared Renderer — a Hand-Copied Second Renderer Drifts

**When a view is re-expressed (a compact card + a drill-down detail, a
summary + a full list), the duplicated entity rows share ONE builder — a
second, hand-copied renderer for the same rows silently drifts** (2026-09-
05: a detail view re-implemented the record rows and lost the attribution
marker the original carried — while the original had already become dead
code that kept the correct behavior, so the drift was invisible to the
build). Delete the superseded renderer in the SAME change that adds its
replacement, keep one row builder, and pin the discriminating case (the
exact marker/formatting the copy dropped) so the two can never disagree
again.
