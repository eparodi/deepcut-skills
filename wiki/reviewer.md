# reviewer

> **Generated** from `.agents/skills/reviewer/SKILL.md` — do not edit by hand.
> Source: [SKILL.md](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md)

Code Reviewer — reviews PRs for correctness, style, security, and adherence to project standards. Produces structured severity-ranked reports; in Fix Loop Mode drives the findings to zero (fix, evidence, full re-review, repeat) until a fresh review passes. Load for reviews and for "fix the review" / "loop until everything is okay".

**Category:** role · **Tags:** review

## Sections

- [Review Scope](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#review-scope)
- [Review Format](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#review-format)
- [Severity Guide](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#severity-guide)
- [Pre-Review Checklist](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#pre-review-checklist)
- [Automatic Trigger](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#automatic-trigger)
- [Fix Loop Mode — drive the findings to zero](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#fix-loop-mode--drive-the-findings-to-zero)
  - [The loop](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#the-loop)
  - [Loop table (keep current)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#loop-table-keep-current)
  - [Hard rules](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#hard-rules)
  - [Anti-patterns (observed in real sessions)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/reviewer/SKILL.md#anti-patterns-observed-in-real-sessions)

