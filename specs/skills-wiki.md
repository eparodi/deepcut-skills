# Skills Wiki: Generated Catalog + Drift-Proofing

**Status:** Draft
**Owner:** Eliseo
**Created:** 2026-08-26

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
renders. The wiki repo does **not exist yet**: GitHub creates it lazily, so
one-time initialization (create a first page via the wiki tab) is a
prerequisite for publishing (verified 2026-08-26: `git ls-remote
…deepcut-skills.wiki.git` → `Repository not found`).

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

## Decision Points

[NEEDS CLARIFICATION: DP-4 Publish topology]
How generated content reaches the wiki: (a) generate into a committed
in-repo `wiki/` mirror, then a publish script pushes it to
`deepcut-skills.wiki.git` (the stale-check stays fully offline, matching the
family's no-network-in-tests rule; the push is the only network touch) vs
(b) generate directly into a clone of the wiki repo (single copy, no
mirror, but tests depend on the clone existing).
**Recommendation: (a)** — the check must pass without network or external
clones; publishing is one explicit command.

[NEEDS CLARIFICATION: DP-1 Toolchain]
Go (stdlib-only; adds `go.mod` to `skills-test`; the stale-check is a
`go test`) vs Python3 (no `go.mod`; stale-check is a script).
**Recommendation: Go** — `go test` is the repo family's verification bar,
and stdlib-only keeps the "no new dependencies" rule.

[NEEDS CLARIFICATION: DP-2 Catalog scope]
Hub skills only (19) vs all three repos (23, requires scanning sibling
worktree paths or a hand-maintained manifest for per-repo-only skills).
**Recommendation: hub only in v1.**

[NEEDS CLARIFICATION: DP-3 Catalog metadata]
Categories (role/stack/process) and tags aren't in SKILL.md frontmatter.
Options: a small hand-maintained catalog manifest that the check validates
(every skill must have an entry; every entry must resolve — additions fail
the check until the manifest is updated) vs no categories (plain
alphabetical index).
**Recommendation: manifest with validation.**

## Proposed Approach (sketch)

- `go.mod` (module-scoped, no external dependencies)
- `tools/wiki-gen/main.go` — reads `.agents/skills/*/SKILL.md` frontmatter
  + headings and `wiki/catalog.json` metadata; writes the wiki pages
  (index + per-skill pages; layout pending DP-4)
- `tools/wiki-gen/main_test.go` — `TestWikiUpToDate` (regenerate to a temp
  dir, diff), catalog-completeness test, frontmatter-parseability test over
  every current SKILL.md
- Publish step — pushes the generated pages to `deepcut-skills.wiki.git`
  (requires the one-time wiki initialization)
- Commands: `go run ./tools/wiki-gen` (regenerate), `go test ./...`
  (verify), publish (push to the wiki remote)
