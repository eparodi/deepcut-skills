# go-htmx

> **Generated** from `.agents/skills/go-htmx/SKILL.md` — do not edit by hand.
> Source: [SKILL.md](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md)

HTMX usage rules for Go html/template server-side apps — the partial-render contract (Rule A/B), required v4 config, attribute patterns, forms/validation, history, security, and testing. Load when writing or reviewing any hx-* attribute, HX-* header, partial render, or htmx wiring.

**Category:** stack · **Tags:** go, htmx, html

## Sections

- [First: read the spec](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#first-read-the-spec)
- [Delivery (how htmx reaches the page)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#delivery-how-htmx-reaches-the-page)
- [The render contract (Rule A/B)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#the-render-contract-rule-ab)
- [Required configuration (set once, in the shared layout)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#required-configuration-set-once-in-the-shared-layout)
- [Attribute rules](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#attribute-rules)
- [Forms & validation](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#forms--validation)
- [Auth & session](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#auth--session)
- [Component JS integration](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#component-js-integration)
- [History rules](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#history-rules)
- [Security](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#security)
- [Testing](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#testing)
- [Common traps (hallucination magnet — verify before trusting)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#common-traps-hallucination-magnet--verify-before-trusting)
- [Cheat sheet (verified against the pinned v4 dist, 2026-09-05)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/go-htmx/SKILL.md#cheat-sheet-verified-against-the-pinned-v4-dist-2026-09-05)

