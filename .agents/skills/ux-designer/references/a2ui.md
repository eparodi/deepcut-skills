# A2UI — Agent-Generated UI (research for the UX Designer)

> Status: research reference for the `ux-designer` role. Researched
> 2026-08-13 from the official sources listed at the bottom. Not a
> feature spec — this document tells the designer HOW agent-generated
> UI changes what we design, and which of our rules become load-bearing.

## 1. What A2UI is

**A2UI (Agent-to-UI)** is an open-source protocol, created by Google
(with CopilotKit contributions), that lets AI agents "speak UI" instead
of speaking only text or code. The agent sends a **declarative JSON
payload** describing the *intent* of a UI (a component tree + a data
model); the client application renders it with its **own native
component library**. The project's own framing:

> "Safe like data, but expressive like code."

- License: Apache 2.0. Repo: `github.com/a2ui-project/a2ui`. Docs: `a2ui.org`.
- Maturity (as researched 2026-08-13): **early-stage public preview.**
  v0.9.1 is the current production release; v1.0 is a release candidate
  (adds client→server RPC via `actionResponse`); v0.8 is legacy. The
  spec is functional but still evolving — expect changes.
- Renderers today: Lit (web), Angular, Flutter (via the GenUI SDK),
  Markdown; **React is on the roadmap** (CopilotKit ships an A2UI
  runtime usable with Next.js; scaffolding via `create-ag-ui-app`).
- Transports: A2A, AG-UI, MCP, SSE/WebSockets.

## 2. Core architecture (the five concepts that matter to a designer)

1. **Surfaces** — a named, catalog-bound render target (think: a panel,
   a card area, a chat-embedded canvas). `createSurface` creates it,
   `deleteSurface` removes it.
2. **Components as a flat adjacency list** — the UI is NOT a nested
   JSON tree. Every component is an object with an `id`, a `type`
   (catalog key), `properties`, and `children` **by ID reference**.
   Flat lists are easier for LLMs to emit incrementally and cheaper to
   patch: `updateComponents` can change ONE component without
   re-sending the tree.
3. **Data model + JSON Pointer binding** — structure is separate from
   state. Component properties can bind to paths in a data model
   (`{path: "/user/name"}`), so state changes (`updateDataModel`) are
   reactive updates, not re-renders.
4. **Message types** (v0.9.1): `createSurface`, `updateComponents`,
   `updateDataModel`, `deleteSurface`. v1.0 adds `actionResponse` for
   client-initiated RPC (e.g., a form submission returning values).
5. **Catalogs** — the client declares which component types an agent is
   ALLOWED to use. An agent can only request components from that
   catalog. Custom components can be registered ("smart wrappers"),
   with sandboxing/trust decisions owned by the developer.

## 3. Why this matters for UX

- **The client enforces the design system.** The agent never styles
  anything: no colors, no fonts, no spacing, no raw CSS. The client's
  catalog mapping (abstract type → native widget) is the ONLY styling
  surface. This is the "tokens, never raw values" rule made structural.
- **Security by design.** Declarative data, not executable code — no
  UI-injection or code-execution risk from agent output.
- **Progressive rendering.** UI streams in and updates in place as the
  agent works — the user watches the interface build instead of waiting
  for a complete response.
- **One payload, many clients.** The same JSON renders on web, mobile,
  desktop with each client's native look.

## 4. What the designer OWNS in an A2UI world

The designer's job shifts from "design each screen" to "design the
catalog and its rules" — the agent becomes a composer, the designer is
the one who decides what may be composed:

1. **The catalog IS the design system.** Component names, properties,
   variants, and states are design decisions. A catalog entry must read
   like our component spec format: purpose, variants, states,
   props/types, responsive behavior, accessibility, tokens.
2. **Design states, not screens.** The agent picks the composition; the
   designer guarantees that every component handles its states
   (loading / empty / error / populated / disabled) — because the agent
   will switch states via `updateDataModel`, not by swapping screens.
3. **Theming is client-side by contract.** A2UI v1.0 renames theme to
   `surfaceProperties`; agents describe semantics, never pixels. Our
   token roles (`--color-*`, `--space-*`, `--text-*`) are exactly the
   surface the agent must NOT reach past.
4. **Accessibility lives in the renderer, so it must be designed into
   the catalog.** ARIA roles, labels, keyboard behavior, focus order,
   and contrast are catalog-component properties. If a component ships
   without them, EVERY agent-generated screen that uses it inherits the
   defect.
5. **Responsive behavior is a catalog contract.** Agents compose with
   layout components (row/column/list), never absolute positioning; the
   designer defines how each catalog component adapts per breakpoint.
6. **Progressive/partial states are first-class.** Streaming
   `updateComponents` means users see partially built UIs — skeletons,
   partial forms, and "still generating" indicators must be designed as
   normal states, not edge cases.
7. **Safety is a design surface.** The catalog allowlist decides what
   an agent MAY show. Anything with money or destructive consequences
   needs confirmation components, disabled-while-pending states, and
   clear consequence copy (our hard rule #14 applies verbatim).

## 5. Where A2UI fits OUR products

**deepcut-live (Next.js frontend):**
- Good candidates: an agent/assistant panel inside the stream or
  dashboard (generated summaries, channel setup wizards, VOD
  highlight pickers, bespoke forms for stream settings), notification
  and mod-card surfaces.
- Keep OUT of agent-generated surfaces: auth/sign-in, the stream key
  and regeneration flow, force-end controls, the video player, chat
  send — anything where a wrong composition is high-cost. Those stay
  hand-built deterministic React.
- Stack note: no official React renderer yet (roadmap); CopilotKit's
  A2UI runtime with Next.js is the practical pilot path today. Do not
  add it as a dependency without a spec + user approval (AGENTS.md).

**deepcut-binance-bot (Go dashboard):**
- Good candidates: agent-generated plain-language explanations of
  trades/cycles rendered as catalog components (cards, charts, tables)
  inside the existing static pages; "explain this loss" panels.
- Hard constraints: money numbers keep the Financial Analyst's display
  semantics (the agent may CHOOSE components, never re-format money);
  red/green pairing, freshness timestamps, and "simulated/real"
  vocabulary remain client-side; no write/trading controls via agent
  UI, ever (dashboard stays read-only).

## 6. Rules we adopt NOW (before any renderer exists)

These make our specs A2UI-ready and cost nothing today:

1. Every component spec gets a **catalog-style props schema** (names,
   types, required/optional, and which API field feeds each prop) —
   our component spec format already does 80% of this; make the props
   table JSON-serializable.
2. Every component's **states must be data-bound states** (one
   component + state fields), never alternate bespoke screens.
3. **Tokens only.** A component property may reference a token role
   (`color: "danger"`), never a raw value — this is what keeps
   agent-composed UI on-design-system.
4. Interactive components declare their **action semantics**: what the
   interaction means, what data it sends, and its disabled/confirm
   variants (A2UI `actionResponse` maps to this).
5. Design **partial/progressive states** for any surface that will be
   generated or streamed.
6. **Critical flows stay hand-built.** Agent UI augments; it does not
   replace deterministic UI where failure is expensive.

## 7. Evaluation summary for the team

| Question | Answer (as of 2026-08-13) |
|---|---|
| Is it production-ready? | Early-stage public preview; v0.9.1 stable, v1.0 RC. Adopt the CONCEPTS now, the runtime later |
| Renderers | Lit, Angular, Flutter (GenUI), Markdown; React on roadmap; CopilotKit runtime for Next.js |
| Transport | A2A, AG-UI, MCP, SSE/WS — our Next.js proxy + WS patterns map to AG-UI/SSE |
| Risk | Spec churn between versions; no official React renderer yet |
| Recommendation | Catalog-first design rules now; pilot an assistant-panel surface with Lit or CopilotKit AFTER a spec is approved; never on money/streaming-critical flows |

## Sources

- A2UI project README: `https://github.com/google/A2UI`
- Official docs: `https://a2ui.org/` (Concepts → Overview, Components
  & Structure, Data Binding, Catalogs, Actions, Theming & Styling)
- Message reference: `https://a2ui.org/reference/messages/`
- Component gallery: `https://a2ui.org/reference/components/`
- Flutter GenUI SDK: `https://github.com/flutter/genui`
- CopilotKit A2UI runtime: `https://docs.copilotkit.ai/generative-ui/a2ui`
