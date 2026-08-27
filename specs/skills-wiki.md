# Skills Wiki: Generated Catalog + Drift-Proofing

**Status:** Implemented
**Owner:** Eliseo
**Created:** 2026-08-26
**Updated:** 2026-08-26 (requirements + design approved at gates; implemented + published)

## Context

`skills-test` is the shared hub for agent skills (`.agents/skills/*/SKILL.md`
— currently 19 skills: 15 role agents, 3 stack skills, 1 process skill).
The hub also carries `AGENT_INDEX.md`, a hand-maintained roster that has
already drifted: its header says "Last updated: 2026-08-15", while the
2026-08-25 scrub changed every SKILL.md it catalogs.

The goal: a wiki for the project whose content is **generated** from the
SKILL.md files — the files stay the single source of truth — so that "every
update we make on a skill, we update the wiki" is enforced by a failing
check instead of by a habit. This follows the repo family's hard-won
lesson: documented-but-unenforced rules diverge (deepcut-binance-bot
AGENTS.md §10.15; skills-test AGENTS.md §10.14).

The wiki is the **GitHub wiki** of this repo
(`github.com/eparodi/deepcut-skills/wiki`, git remote
`deepcut-skills.wiki.git`) — a separate git repository whose pages GitHub
renders. **Initialized 2026-08-26** (user created the first page): remote
is live with a `master` branch (HEAD `40055d1`); the placeholder `Home.md`
("Welcome to the deepcut-skills wiki!") is replaced by the generated home
page on first publish.

## Requirements

### User Story 1: Skill catalog (discoverability)

As a user, I want a browsable wiki that catalogs every skill in the hub, so
I can discover what skills exist, what each covers, and where each lives.

**Acceptance Criteria:**
- Given the current hub skill set, When I open the wiki's home page
  (`Home.md`), Then it lists every skill in `.agents/skills/` with its
  one-line description and a link to its page.
- Given any skill in `.agents/skills/`, When I open its wiki page, Then it
  shows the skill's name, description (frontmatter), source file location,
  and section outline (the `##` headings of its SKILL.md).
- Given the generated wiki, When I compare it to `.agents/skills/`, Then no
  skill is missing from the index and no page references a skill that
  doesn't exist (no dead links).

### User Story 2: Single source of truth (no drift)

As a maintainer, I want the wiki generated from the SKILL.md files, so the
wiki can never silently diverge from the skills.

**Acceptance Criteria:**
- Given the current skills, When I run the generation command, Then
  `wiki/` is rewritten and no manual edits to generated pages are required.
- Given any change to a skill (name, description, headings, or a
  new/removed skill), When I run the verification check without
  regenerating, Then it fails with a non-zero exit and names the stale
  files.
- Given unchanged skills, When I run generation twice, Then the second run
  produces byte-identical output (deterministic — no timestamps or
  nondeterministic content).
- Given a freshly regenerated wiki, When I run the verification check, Then
  it passes.
- Given any generated wiki page, Then it carries a visible "generated from
  <source path>; do not edit by hand" marker, so generated content is never
  hand-edited.
- Given regenerated content, When I run the publish step, Then the GitHub
  wiki shows the same pages (a push to `deepcut-skills.wiki.git` succeeds
  and the wiki renders the catalog).

### User Story 3: Enforcement on skill changes

As a maintainer, I want the stale-check wired into the skill-change
workflow, so "every skill update → wiki update" is enforced, not promised.

**Acceptance Criteria:**
- Given a new skill added to `.agents/skills/`, When the verification check
  runs, Then it fails until the wiki is regenerated (and catalog metadata
  updated if required).
- Given catalog metadata changes (e.g., a skill's category), When the
  verification check runs, Then it fails until the index is regenerated.
- Given the skill-lifecycle docs (skill-factory, learning-porter, the retro
  template in `specs/memories/README.md`), When they describe creating or
  updating a skill, Then they reference the one-command regenerate + publish
  flow.

## Non-Goals

- ❌ Self-hosted or in-repo-only wiki — the wiki IS the GitHub wiki; any
  in-repo `wiki/` directory is only the generated build artifact (pending
  DP-4), never the wiki itself.
- ❌ Manual wiki page edits (GitHub UI or by hand) — every page is generated
  from the SKILL.md files and pushed by the publish step.
- ❌ Hand-written pages duplicating SKILL.md bodies — the catalog is
  generated; the SKILL.md stays the single source of truth.
- ❌ Cross-repo catalog of per-repo-only skills (`bot-engineer`,
  `ui-engineer`, `go-bot`, `deepcut-platform`) in v1 — `AGENT_INDEX.md`
  remains the cross-repo roster (see DP-2).
- ❌ CI wiring or git hooks — `skills-test` has no CI today; enforcement is
  the local check command (adding CI is a follow-up).
- ❌ A hosted website — the wiki renders in the repo's markdown viewer.
- ❌ Replacing retros, session logs, or `AGENTS.md` §10 — the wiki catalogs
  skills, not learnings.

## Resolved Decisions (review gate, 2026-08-26)

1. **DP-1 — Toolchain: Go, stdlib-only.** New `go.mod` (module
   `deepcut-skills`), tool at `tools/wiki-gen/`. The stale-check is a Go
   test: `go test ./...` from the repo root. No external dependencies.
2. **DP-2 — Scope: hub skills only.** The catalog covers the 19 skills in
   `skills-test/.agents/skills/`. Per-repo-only skills (`bot-engineer`,
   `ui-engineer`, `go-bot`, `deepcut-platform`) stay in `AGENT_INDEX.md`
   (follow-up: extend the wiki when a cross-repo view is wanted).
3. **DP-3 — Catalog metadata: `wiki/catalog.json` manifest, validated.**
   Bidirectional check: every skill dir must have an entry; every entry
   must resolve; `category` ∈ {role, stack, process}. The manifest doubles
   as the tripwire — adding a skill fails the check until it's cataloged.
4. **DP-4 — Publish topology: in-repo mirror + push script.** Generated
   pages are committed under `wiki/` (offline-verifiable, versioned with
   the skills); `tools/wiki-gen/publish.sh` pushes them to
   `deepcut-skills.wiki.git` (branch `master`).
5. **Wiki initialized 2026-08-26.** Remote live (`master` @ `40055d1`);
   placeholder `Home.md` replaced on first publish.

## Design (added 2026-08-26, pending gate)

### Architecture

**Files**

| Path | Purpose |
|------|---------|
| `go.mod` | Module `deepcut-skills`, stdlib only |
| `tools/wiki-gen/main.go` | CLI: default regenerates `wiki/`; `--check` regenerates to a temp dir and exits 1 when stale |
| `tools/wiki-gen/generate.go` | Core: `loadSkills`, `loadCatalog`, page renderers, `generate` |
| `tools/wiki-gen/generate_test.go` | Pinned checks (below) |
| `tools/wiki-gen/publish.sh` | Publish flow (below) |
| `wiki/catalog.json` | Manifest: skill → {category, tags, notes} |
| `wiki/` | Generated output, committed: `Home.md`, `_Sidebar.md`, `<name>.md` ×19 |

**Frontmatter parsing** — within `---` fences, `name:` and `description:`
are consumed; value = text after the first `:`, trimmed, optional
surrounding double quotes stripped, trailing `\r` stripped. `name` must
equal the skill directory name; missing/empty values are errors (pinned
by `TestFrontmatterParses`).

**Determinism** — output depends only on the skill files, the catalog, and
the templates: skills sorted by name, no timestamps, no randomness.
Byte-identical regeneration is pinned by `TestDeterministic`.

**Generated pages**

- `wiki/Home.md` — index: intro, "generated; do not edit" marker, catalog
tables grouped by category (role agents / stack skills / process
skills), per-skill link + one-line description, and a
"Regenerating & publishing" section with the commands.
- `wiki/_Sidebar.md` — GitHub-wiki sidebar: category-grouped link list.
- `wiki/<name>.md` — per-skill page: name, generated-marker with the
source path, source link to the main repo's blob URL
(`github.com/eparodi/deepcut-skills/blob/main/...`), frontmatter
description, category/tags/notes, and a "Sections" outline: the
SKILL.md's `##` headings with their `###` children, each linking to its
GitHub blob anchor (duplicate headings get GitHub's `-1`, `-2` suffixes;
headings inside fenced code blocks are excluded so template examples
never leak). Omitted when the skill has no `##` headings.

**Pinned checks (`tools/wiki-gen/generate_test.go`)**

| Test | Pins |
|------|------|
| `TestCatalogCompleteness` | Manifest ↔ skill-dir coverage (bidirectional), category enum |
| `TestFrontmatterParses` | Every SKILL.md: name == dir name, non-empty description |
| `TestWikiUpToDate` | Regenerate to temp; tree diff vs `wiki/`; fails naming stale/missing/extra files |
| `TestDeterministic` | Two regenerations are byte-identical |
| `TestHomeLinksResolve` | Every link target on the home page resolves to a generated page (no dead links) |
| `TestPagesCarryMarker` | Every generated page carries the "do not edit" marker |
| `TestSlugify` | GitHub anchor algorithm (verified against live blob pages) |
| `TestSectionsSkipCodeBlocks` | Headings inside fenced code blocks never appear in the outline |
| `TestNoTemplateLeak` | Template-embedding skills produce exactly their real section count |

**Publish flow (`tools/wiki-gen/publish.sh`)**

1. `go run ./tools/wiki-gen` — regenerate in place.
2. `go test ./...` — must pass (the wiki is current).
3. Clone `deepcut-skills.wiki.git` into a temp dir.
4. Copy the generated `wiki/*.md` files over the clone's working tree
   (add/update only).
5. List wiki pages that are NOT in the generated set — print them, do NOT
   delete (the wiki may carry hand-authored pages; removing is a manual
   decision).
6. Commit `chore: update wiki (generated from skills)` and push `master`.
7. Remove the temp clone.

**Commands**

| Command | Action |
|---------|--------|
| `go run ./tools/wiki-gen` | Regenerate `wiki/` |
| `go run ./tools/wiki-gen --check` | Stale check (exit 1 when stale) |
| `go test ./...` | Full verification (the enforced check) |
| `./tools/wiki-gen/publish.sh` | Publish to the GitHub wiki |

> Link syntax note: pages use plain relative markdown links
> (`[pm](pm)`); GitHub-wiki rendering of relative links is verified on the
> first publish, and `[[wiki]]`-style links are the fallback if needed.

## Implementation Notes

- **2026-08-26 (post-approval, operator feedback "the section part is
  quite poor"):** the Sections outline was flat, unlinked, and leaked
  template headings. Fix: (1) `headingsOf` → `sectionsOf`, which skips
  fenced code blocks — skill-factory, learning-porter, qa,
  security-engineer, and reviewer embed template examples whose `##`
  headings previously polluted their outlines (skill-factory showed 12
  sections, 6 of them from a code block); (2) the outline is now nested
  (`##` with `###` children) and each item links to its GitHub blob
  anchor via a github-slugger-compatible `slugify` + duplicate `-1`/`-2`
  suffix handling, verified against the live blob pages (spec-driven's
  anchors match byte-for-byte, including `process`/`process-1`/
  `process-2` and `review-gate`/`review-gate-1`). Pinned by
  `TestSlugify`, `TestSectionsSkipCodeBlocks`, `TestNoTemplateLeak`.

- **2026-08-26 (post-merge follow-up, DP-2):** the catalog now covers the
  per-repo-only skills (`bot-engineer`, `ui-engineer`, `go-bot`,
  `deepcut-platform`) via a `per_repo` manifest section. Rationale: hub
  copies would carry repo-specific references, violating §10.48 (hub free
  of external project references), and scanning sibling worktrees would
  break the offline stale-check. Instead the manifest mirrors each skill's
  one-line description (copied exactly from the source frontmatter at
  write time), validates name/category/repo/source-path offline, and each
  generated page links to the real SKILL.md blob in its own repo. The
  learning-porter owns keeping the mirrored descriptions in sync.

## Task Checklist (Phase 3)

1. [x] (Tooling) Scaffold `go.mod`; write `tools/wiki-gen/generate_test.go`
   (pinned checks that don't need the manifest)
   → Test: `TestFrontmatterParses`, `TestDeterministic` (red first)
   → Satisfies: US2 (generation, determinism)
2. [x] (Tooling) Implement `tools/wiki-gen/generate.go` + `main.go`
   (regenerate + `--check`)
   → Satisfies: US2 AC1–AC5, US1 AC2
3. [x] (Catalog) Create `wiki/catalog.json` for all 19 skills
   → Test: `TestCatalogCompleteness`
   → Satisfies: US3 AC1–AC2 (the tripwire), US1 AC1
4. [x] (Output) Generate `wiki/` (Home.md, _Sidebar.md, 19 pages); wire
   `TestWikiUpToDate`, `TestHomeLinksResolve`, `TestPagesCarryMarker`
   → Satisfies: US1 AC1–AC3, US2 AC2–AC5
5. [x] (Publish) Write `tools/wiki-gen/publish.sh`
   → Satisfies: US2 AC6 (publish AC)
6. [x] (Docs) Wire skill-factory, learning-porter, `specs/memories/README.md`
   to reference the regenerate + publish flow
   → Satisfies: US3 AC3
7. [x] (Live) Run `publish.sh`; verify the wiki renders (link syntax check)
8. [x] (PR) Open PR `feat/skills-wiki` → main (#28)
