# AI Engineer Role — Research Sources

Gathered 2026-08-13 to seed the `ai-engineer` skill (template version
lives in `skills-test`; the reference implementation is
`deepcut-binance-bot`). Goal of the role:
improve the DeepSeek connection — (1) handle malformed responses better and
(2) reduce the number/cost of requests to DeepSeek.

## Where the pain is today (codebase facts)

- `internal/llm/deepseek.go`: single-shot POST to `/chat/completions`,
  60s HTTP client timeout, **no retry, no backoff, no cache, no
  response_format**, body read capped at 1 MiB. Non-200 → error, empty
  choices → error.
- `internal/ensemble/agent.go`: malformed persona JSON → HOLD + warning
  (spec-mandated, correct behavior — but no repair attempt, no one-shot
  re-prompt, no fenced-codeblock extraction).
- `internal/ensemble/ensemble.go`: every cycle = 3 persona calls per
  symbol (N symbols × 3 × 1 per 15m), plus reflection calls hourly.
  `MaxTokens: 800`, `llmTimeout: 60s`, parallel per symbol.

## DeepSeek platform facts (verified from api-docs.deepseek.com, 2026-08-13)

- **OpenAI-compatible** protocol (`https://api.deepseek.com`), models
  `deepseek-v4-flash` / `deepseek-v4-pro`. Flash = cheap tier, Pro = quality
  tier. New `thinking: {"type": "enabled"}` and `reasoning_effort` params
  exist on v4 models (cost/quality tradeoff knob).
- **JSON Output mode exists**: `response_format: {"type": "json_object"}`
  + the word "json" + a format example in the prompt + a `max_tokens` large
  enough to avoid truncation. **Known issue (documented by DeepSeek): the
  API occasionally returns empty content in JSON mode** — the client must
  tolerate that (→ treat as malformed, not as transport error).
- **Context caching is automatic and enabled by default** (disk-based KV
  cache). Hits require a request prefix that **fully matches a previously
  persisted cache-prefix unit**. Response `usage` carries
  `prompt_cache_hit_tokens` / `prompt_cache_miss_tokens` → measurable.
  Best-effort, expires within hours–days. Practical consequence: put the
  **stable part of the prompt first** (system prompt + shared market
  context), per-symbol/per-cycle variable data last.
- **Error codes**: 400 invalid format, 401 auth, 402 insufficient balance,
  422 invalid parameters, 429 rate limit, 500 server error, 503 overloaded.
  429/5xx are retryable; 400/401/402/422 are not.
- **Concurrency limits**: 500 (pro) / 2500 (flash) concurrent requests per
  account; exceeding → HTTP 429. Optional `user_id` param for isolation.
- **Keep-alive quirk**: non-streaming responses may contain empty lines and
  streaming responses SSE comments (`: keep-alive`) — a hand-rolled parser
  must tolerate them. Server closes the connection if inference hasn't
  started within 10 minutes.
- Pricing page did not load via automated fetch — cache-hit discount rates
  NOT captured. Do not cite numbers.

## Reference repos (verified via GitHub API, 2026-08-13)

| Repo | Stars | What it teaches this bot |
|---|---|---|
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~56.3k | The reference design for LLM gateways: retries with exponential backoff, fallbacks across models/providers, caching layers, rate-limit awareness, per-request cost tracking. Read their retry/fallback docs for the exact taxonomy of "which errors to retry". |
| [567-labs/instructor](https://github.com/567-labs/instructor) | ~13.7k | Structured-output library. Key ideas: validate the parsed object against a schema, and **automatically re-prompt once with the validation error** before failing. This is the canonical malformed-response recovery loop. (formerly instructor-ai/instructor) |
| [BoundaryML/baml](https://github.com/BoundaryML/baml) | ~9.0k | Declarative function schemas + explicit **retry policies** (`RetryPolicy` with per-failure-type behavior) for LLM calls. Good mental model for config-driven retry in the bot. |
| [promptfoo/promptfoo](https://github.com/promptfoo/promptfoo) | ~24.2k | LLM **eval/regression testing**: run prompt variants against pinned inputs, assert output shape. The right pattern for the persona-prompt test suite (and it explicitly supports DeepSeek). |
| [openai/openai-cookbook](https://github.com/openai/openai-cookbook) | ~75.2k | Official OpenAI patterns: rate-limit handling with backoff, prompt caching, structured outputs, "how to count tokens first" (estimate and truncate history). DeepSeek is protocol-compatible, so most recipes transfer. |
| [anthropics/claude-cookbooks](https://github.com/anthropics/claude-cookbooks) | ~51.5k | Prompt-caching best practices (static prefix → dynamic suffix), JSON mode usage, multi-turn context management. |
| [vercel/ai](https://github.com/vercel/ai) | ~26.2k | TypeScript AI SDK; useful mainly for its error taxonomy and streaming/retry semantics rather than code to copy. |
| [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) | ~10.7k | Go client for the OpenAI-compatible protocol. A potential dependency *if* we ever move off the stdlib hand-rolled client — see constraint below. |
| [openai/openai-go](https://github.com/openai/openai-go) | ~3.4k | Official Go SDK (Apache-2.0). Built-in retries (`WithMaxRetries`), 429/5xx handling, `ResponseFormatJSONObject`. Strong candidate if a dependency is approved. |
| [josdejong/jsonrepair](https://github.com/josdejong/jsonrepair) | ~2.4k | Algorithm reference for repairing broken JSON (unquoted keys, trailing commas, truncation). TypeScript, so NOT a dependency — mine it for the repair rules we reimplement in Go if we add a repair pass. |

## Patterns to adopt (candidate, for the spec phase — not decided yet)

1. **Malformed-response ladder** (priority order, all config-driven):
   a. `response_format: json_object` at the API level (fewer malformed
      responses to begin with).
   b. Cheap local repair: strip markdown fences, extract first balanced
      JSON object, tolerate trailing commas.
   c. One-shot re-prompt with the validation error appended (instructor
      pattern), with a per-cycle retry budget (e.g. 1) and only on
      transport/429/5xx/malformed — never on 4xx business errors.
   d. Fallback unchanged: HOLD + warning (spec-mandated).
2. **Request reduction** (each changes behavior → needs spec sign-off):
   - Cache-friendly prompt layout (stable prefix first) to maximize
     DeepSeek's automatic KV-cache hits — reduces *cost*, not request count.
   - Skip re-asking personas when inputs haven't changed since last cycle
     (e.g. symbol in cooldown, no new closed candle, unchanged indicators)
     → reuse the persisted decision with a stale flag.
   - Model tiering: flash for low-signal/reflection calls, pro for
     decision calls (or config-driven per persona — already possible).
   - Batch multiple symbols into ONE request returning a JSON map
     (symbol → decision) — trades 3N calls for N, at higher malformed-blast
     radius. Needs careful trade-off analysis.
   - Reduce reflection frequency or cap tokens further.
3. **Transport hardening**: exponential backoff + jitter on 429/500/503
   (respecting `Retry-After` if present), circuit breaker after N
   consecutive failures (skip personas, HOLD, log), 402 balance
   detection → loud alarm, and tolerant parsing of keep-alive lines.
4. **Observability**: log `prompt_cache_hit_tokens` / miss tokens, per-request
   latency, per-persona malformed-rate and retry-rate → feed the dashboard.
   This is what makes "did our changes help?" answerable.

## Constraints from this repo

- AGENTS.md forbids new dependencies without explicit user approval —
  treat openai-go as a *candidate*, not a plan.
- The four interfaces (`Provider`, `DecisionMaker`, `MemoryStore`,
  `Sampler`) are the spec contract; all of the above must sit behind
  `Provider`/`DecisionMaker`, never change their signatures unilaterally.
- Malformed → HOLD is spec-mandated behavior (US3). A retry ladder changes
  request counts → spec change, not silent code change.
- Test data must use obviously-fake API keys (`test_`, `fake_`).

## Open questions to resolve with the user (spec phase)

1. Is 1 retry on transport error acceptable (request-count cost vs resilience)?
2. Is one-shot re-prompt on malformed JSON acceptable?
3. Which request-reduction options are in scope (skip-unchanged, tiering,
   batching)?
4. Is adding the official Go SDK (openai-go) worth it vs hardening the
   stdlib client?
