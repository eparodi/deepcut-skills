# Retro: Agent Autonomy Rules — 2026-08-10

## What Happened

The agent workflow had three friction points:
1. **Features branched from stale branches** — no rule saying "always branch from `main`"
2. **Zed prompted for every file operation** — `move_path` needed confirmation, AGENTS.md didn't assert write autonomy
3. **Skills-test AGENTS.md diverged from deepcut-live** — different section numbering caused merge conflicts

## Root Cause Analysis

### Correction #1: No base-branch rule
- **Missing rule:** AGENTS.md §6.2 (Branching Convention) listed prefixes and naming but never specified *which branch to start from*
- **Fix:** Added explicit "always branch from latest `main`" with fetch-checkout-pull sequence
- **Prevention:** The branching convention section now has a mandatory pre-branch checklist

### Correction #2: File-operation prompts
- **Missing rule (Zed):** `move_path` was the only tool still set to `"confirm"` — inconsistent with `write_file`, `create_directory`, `delete_path`, `copy_path` all set to `"allow"`
- **Missing rule (AGENTS.md):** §7.2 listed individual allowed tools instead of granting blanket write autonomy with a blocklist
- **Fix:** Changed `move_path` to `"allow"` in Zed config + rewrote §7.2 to "Fully Autonomous Within the Repo" with explicit destructive-op blocklist
- **Prevention:** New AGENTS.md rule explicitly states: "You have full permission to create, edit, and delete any file inside the repository working tree. Do not ask for confirmation — Git is your safety net."

### Correction #3: Repos drifted apart
- **Missing mechanism:** No CI check or process to keep AGENTS.md aligned between `deepcut-live` and `skills-test`
- **Fix:** Resolved merge conflict manually
- **Prevention:** Recommended (not implemented) — a CI job that diffs the two AGENTS.md files and warns on divergence

### Correction #4: git-worktree edge case
- **Not a rule gap, but a tool limitation:** `git checkout main` fails when `main` is already checked out in another worktree
- **Workaround:** `git -C <primary-worktree> pull` or `git fetch origin main:main`
- **Note:** The new §6.2 branching instructions are correct as *intent* but the agent will discover this edge case when it tries to `git checkout main` inside a secondary worktree

## Rule Updates Made

### deepcut-live/AGENTS.md (PR #16)

| Section | Before | After |
|---|---|---|
| §6.1 | "Never Commit Without Explicit Request" | "Commit & Push" — auto on feature branches, gated on shared |
| §6.2 | No base-branch rule | Mandatory fetch-checkout-pull from `main` |
| §7.2 | List of allowed tools | "Fully Autonomous Within the Repo" + blocklist |

### skills-test/AGENTS.md (PR #7)
Same changes + retained upstream `gh pr create` backtick tip.

### ~/.config/zed/settings.json
`move_path.default`: `"confirm"` → `"allow"`

## Open Issues

1. **Should we add a CI sync check?** A job that diffs `AGENTS.md` between repos and flags divergence would prevent future merge conflicts.
2. **Should the branching rule mention worktrees?** The "checkout main" step needs a worktree-aware alternative. Could add: "If `main` is checked out in another worktree, use `git fetch origin main:main` instead."
3. **Should agent profiles be project-enforced?** Zed supports `.zed/settings.json` at the project level — this could lock `default_profile: "write"` for all collaborators.

---

*Cross-referenced from: [2026-08-10-session-log.md](./2026-08-10-session-log.md)*
