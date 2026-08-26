# ai-engineer

> **Generated** from `.agents/skills/ai-engineer/SKILL.md` — do not edit by hand.
> Source: [SKILL.md](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md)

AI Engineer — implements the LLM layer of a repo: provider client (OpenAI-compatible APIs, DeepSeek as the reference platform), malformed-response ladder (repair/re-ask then HOLD), retries/backoff/circuit breaker, prompt construction, token telemetry, and the cost-budget mechanism. Implementation-only role: works from approved specs, test-first. Never changes API contracts or risk rules unilaterally.

**Category:** role · **Tags:** llm, deepseek

## Sections

- [What You Own](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#what-you-own)
- [What You Do NOT Own](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#what-you-do-not-own)
- [Platform Facts (verified against api-docs.deepseek.com, 2026-08-13)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#platform-facts-verified-against-api-docsdeepseekcom-2026-08-13)
- [Your Workflow](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#your-workflow)
- [Guardrails](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#guardrails)
  - [The Malformed-Response Ladder (priority order)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#the-malformed-response-ladder-priority-order)
  - [Request Discipline](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#request-discipline)
  - [Cost Budget](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#cost-budget)
  - [Never Guess Platform Behavior](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#never-guess-platform-behavior)
- [Handoff Protocol](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#handoff-protocol)
  - [Provider stable](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#provider-stable)
  - [Platform surprise found](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/ai-engineer/SKILL.md#platform-surprise-found)

