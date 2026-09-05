---
name: go-htmx
description: "HTMX usage rules for Go html/template server-side apps — the partial-render contract (Rule A/B), required v4 config, attribute patterns, forms/validation, history, security, and testing. Load when writing or reviewing any hx-* attribute, HX-* header, partial render, or htmx wiring."
---

# Go + HTMX Standards

HTMX turns `html/template` pages into a partial-swap SPA shell: the
server renders a full document on direct loads and only the swap
region on `HX-Request` requests. This skill is the implementation
rulebook for that architecture. It exists because HTMX is declarative:
the library is small, but its defaults interact with Go handlers,
cookies, caches, and history in non-obvious ways — and **v4 is an
early major whose dist was rewritten (36KB vs v2's 51KB)**: several v2
config keys, attributes, and event names are gone or renamed. Most
rules here come from htmx.org's docs crossed with **direct
verification against the pinned dist** — when the two disagree, the
dist wins (the docs website lags releases).

## First: read the spec

The contract for a host repo's migration (phases, swap target,
non-goals) lives in that repo's spec. This skill implements the
generic rules; it never changes the spec. If the spec and this skill
disagree, the spec wins and this skill gets fixed.

Boundary with other skills in a multi-role repo: the UI/component
skill owns the component registry, CSS, and the enhancement bundle
(htmx attributes in component HTML are yours; component design is
theirs); the Go stack skill owns testing/layout rules (apply them to
htmx handler tests). Declare the boundary explicitly in the host
repo's skills.

## Delivery (how htmx reaches the page)

1. **No CDN, no committed copy.** A small build-time fetch tool
   downloads `htmx.min.js` for an exact pinned release, verifies the
   SHA-256 against a constant, and writes it to a gitignored embed
   directory (`//go:embed` serves it from the binary). Build/deploy
   entry points run the fetch first; a bare `go build` without it
   fails loudly (empty embed match) — that failure is by design,
   document it, never "fix" it by committing the file.
2. **Pin version + checksum from the official GitHub tags — never the
   docs site, never guess.** The docs website can lag the release line
   (2026-09: the docs' examples still showed 2.x while v4.0.0 was
   current). **v4 is an early major** — verify every config key,
   attribute, and event name you rely on against the actual dist
   before shipping it (`grep` the minified file; remember v4 derives
   `HX-`-prefixed response-header names at runtime, so a literal grep
   miss is evidence to double-check, not proof of absence). Never mix
   1.x/2.x/4.x rules — each migration changed defaults and names.
3. One `<script src="/static/htmx.min.js" defer>` tag in the shared
   layout, served by a handler that mirrors the existing JS handler
   (Content-Type, no-store). Pin its presence on every page — a
   controller shipped in a shared bundle does nothing on a page that
   fails to load the bundle.
4. A new htmx-returning endpoint is still a normal route: declare any
   new template in the page registry or the components contract —
   undeclared template files should panic at startup by design.

## The render contract (Rule A/B)

1. The ONE place this lives: a single execute/render choke point in
   the handler layer. Never branch on `HX-Request` anywhere else —
   two branching sites will drift.
2. `r.Header.Get("HX-Request") == "true"` → execute ONLY the page's
   swap-region block (`ExecuteTemplate(w, "<region>", data)`), status
   200, `Content-Type: text/html; charset=utf-8`. Missing/anything
   else → execute the full layout exactly as before. The region block
   is defined in the shared layout (`{{block "<region>" .}}…{{end}}`)
   and covers the WHOLE region the swap target replaces (banner,
   section nav, content, footer). **Never partial-render just the
   content block** (2026-09-05): the swap target usually contains more
   than the content block, so a content-only partial wipes the
   section nav/footer on every swap ("the nav buttons disappear").
   The partial must equal the ENTIRE region the target replaces —
   pin it byte-for-byte against the full page's region in a test.
3. The partial is the swap unit. **Never use `hx-select` on nav links**
   to strip a full page — the server already returns only the
   fragment; double-stripping mechanisms drift. `hx-select` +
   `hx-swap="outerHTML"` is the *alternative* full-page-response
   pattern — legitimate for self-contained widgets, never for page
   navigation.
4. Redirects: htmx requests get **200 + `HX-Redirect` header**, never
   a 302/303 with a body. htmx does not process HX-* response headers
   on 3xx responses (the browser intercepts the redirect first).
   Non-htmx redirects keep 303 for form POSTs.
5. No Post/Redirect/Get for htmx forms — return the updated fragment
   directly (that's the point of partial swaps).
6. `Vary: HX-Request` on responses whose content differs by that
   header (htmx docs rule). Moot while everything is `no-store` — but
   if caching is ever scoped down, this rule is mandatory, not
   optional.

## Required configuration (set once, in the shared layout)

1. `htmx.config.history = "reload"` via the htmx-config meta:
   `{"history":"reload"}`. This is v4's full-page answer to the
   history-cache trap: on a history-cache miss (back/forward, deep
   links) the browser does a FULL reload of the URL — the Rule A full
   document — instead of an AJAX fetch that carries `HX-Request: true`
   and receives a bare partial. It also keeps sensitive data out of
   the client-side history cache by construction (v4's snapshot cache
   is opt-in via the hx-history-cache extension; "reload" never
   snapshots). Do NOT re-add the v2 keys — `historyRestoreAsHxRequest`
   and `historyCacheSize` are GONE and silently ignored by v4's config
   merge (a lying knob).
2. **Validity reporting is built-in** — v4 calls `reportValidity()`
   itself during form-data collection. The v2 `reportValidityOfForms`
   key is gone; do not ship it.
3. **Removed v2 config surface:** `selfRequestsOnly`, `allowScriptTags`,
   `allowEval`, `withCredentials`, `responseHandling`, `timeout` — do
   not reference them. Where a host repo relied on a behavior,
   re-derive it from the dist: scripts in swapped content are still
   processed by default (verified).
4. Config arrives via `<meta name="htmx-config">` (v4 merges it) or JS.
   Pin the meta string in a test.

## Attribute rules

1. **Explicit attributes only — no global `hx-boost`** (host-spec
   non-goal). Boosting swaps the `body` by default, does not update
   `<html>`/`<body>`, and makes per-page pins meaningless.
2. Nav links: `hx-get` + `hx-target="<swap-region>"` +
   `hx-push-url="true"` + `hx-history-elt="<swap-region>"`. The swap
   target is ONE region and ONLY that region — header, sidebar,
   footer nav, auth forms, and floating widgets live outside it and
   must never be re-rendered by navigation. **The same triple applies
   to every IN-PAGE navigation link** that re-renders the swap region
   (filter/symbol pills, pagers, size pills, clear-filters, day
   pickers, cross-page shortcuts, confirm-page back links): a plain
   `<a href>` inside the swap region is a full-page reload that breaks
   the SPA contract (2026-09-05). Full-page-by-design links stay
   attribute-free: preference switchers (cookie + redirect), fragment
   links needing full-page scroll-to-anchor, downloads, JSON API
   links, and the auth flows.
3. Remember hx-* attribute **inheritance**: attributes hoist to parent
   elements and children inherit them. Use `hx-disinherit`/`unset`
   deliberately; do not leave a stray inherited `hx-*` on a parent
   that changes child behavior.
4. Trigger defaults: `input/textarea/select` → `change`, `form` →
   `submit`, everything else → `click`. Modifiers when needed:
   `changed`, `delay:500ms`, `throttle:`, `once`, `from:<selector>`
   (the from-selector is NOT re-evaluated after swap). Polling:
   `hx-trigger="every 2s"`; the v2 "respond 286 to stop polling"
   mechanism is GONE from v4 — verify v4's poll-stop behavior against
   the dist before relying on one.
   **Never put `load` on a fragment that replaces ITSELF**
   (`hx-swap="outerHTML"` back onto its own element): v4 fires a
   `load` trigger IMMEDIATELY at trigger-setup time (verified in the
   dist: the non-load branch adds listeners, the load branch calls the
   handler right away), htmx processes the fresh element after every
   swap, and the trigger fires again — the fragment fetches itself in
   a busy loop at round-trip speed (2026-09-05: a self-refreshing span
   with `hx-trigger="load, every 30s"` issued ~165 requests/second
   through the auth stack). `every Ns` alone is safe: the interval
   checks `isConnected` before firing and self-clears when the old
   element is replaced; the new element re-arms it. If the poller
   element should persist instead, swap `innerHTML` (the
   docs-canonical poll pattern) and return plain content — never
   re-insert the polling attributes.
5. Swap styles: `innerHTML` default (what a swap-region div uses);
   `outerHTML` for self-replacing widgets; modifiers `scroll:top`,
   `show:top`, `ignoreTitle:true` (partials never carry a title to
   swap in — but see the title rule below).
6. Extended target selectors: `this`, `closest <sel>`, `next <sel>`,
   `previous <sel>`, `find <sel>` — prefer these over inventing ids
   for row-level widgets (tables).
7. Racing requests: `hx-sync="closest form:abort"` for input-validation
   vs form-submit races; `htmx:abort` event to cancel programmatically.
8. Request feedback: `htmx-indicator` class (hidden until a
   `htmx-request` class lands on the issuing element), `hx-indicator`
   to point elsewhere, `hx-disabled-elt` to disable submit buttons
   during flight.
9. Fade/transitions: use the built-in `htmx-swapping`/`htmx-settling`
   classes on the swap region (CSS-only — verified in the v4 dist), or
   a single delegated `htmx:after:swap` listener registered ONCE in
   the app bundle — never an inline handler inside swapped content.
10. OOB swaps (`hx-swap-oob="true"` in the response) are the mechanism
    for updating something OUTSIDE the swap region from a partial
    response (e.g., a badge in the header). Table rows need a
    `<template>` wrapper (bare `<tr>` can't stand alone).
    **The nav active-state sync is this pattern's canonical example**
    (2026-09-05): sidebar/bottom navs live outside the swap region,
    so their "here"/aria-current state would never move on navigation.
    The partial renderer appends a small OOB element (a `<meta>` or
    span) carrying the server-computed active hrefs; the app bundle
    applies them on `htmx:after:settle`. The server OWNS the grouping
    — never re-derive it client-side. The static counterpart in the
    layout is the OOB target (an OOB fragment with no existing target
    is dropped).
11. `hx-trigger="load"` is for ONE-SHOT lazy fragments only: a
    placeholder (`hx-get="…" hx-trigger="load" hx-swap="outerHTML"`)
    whose RESPONSE REPLACES it with content that carries NO
    hx-get/hx-trigger — the response must never re-insert the trigger
    (see the busy-loop trap).
12. **Global loading feedback is one request-scoped indicator, not
    per-page loaders** (2026-09-05): a fixed top progress bar (or
    equivalent) driven by the lifecycle — `htmx:before:request` starts
    an in-flight count; a ~150ms grace timer shows it only for
    requests that outlive it (fast pages never flash);
    `htmx:finally:request` (fires on success AND error) decrements and
    hides at zero. **Skip background polls** by source — a periodic
    poll would pulse the bar constantly. The indicator is layout
    chrome (outside the swap region, aria-hidden, above the header in
    the z-stack) and must never appear in a partial.

## Forms & validation

1. **CSRF is unchanged.** A hidden `<input name="csrf">` + the
   existing POST middleware chain keeps working: htmx submits the same
   form body. Never switch CSRF to `hx-headers` — the hidden input is
   the documented preferred mechanism and existing tests pin it.
2. Server-side validation errors: return **200 + partial with the
   error-summary component** (the one pattern). The v2
   `responseHandling` config ("swap 422") is GONE in v4 — never assume
   a 4xx/5xx response swaps or errors; if a status-code pattern is
   ever needed, derive v4's behavior from the dist + the
   `htmx:response:error` event first.
3. `hx-validate="true"` only on non-form elements that need native
   validation. Client validation is UX only — all rules are re-checked
   server-side.
4. Error UX: `htmx:response:error` (HTTP error, renamed from
   `htmx:responseError`) — and the network-error event: VERIFY its
   name against the dist before wiring it (v4 renamed most events to
   colon form: `htmx:afterSwap` → `htmx:after:swap`,
   `htmx:configRequest` → `htmx:config:request`, `htmx:load` →
   `htmx:after:process`). Never assume a v2 event name exists in v4.

## Auth & session

1. **Login, logout, and session-refresh stay full-page. Never give
   them hx-* attributes** (host-spec non-goal). They depend on
   path-scoped cookies and CSRF semantics that partial swaps break —
   a handler on a path the cookie is not scoped to can never read it.
2. Expired session mid-SPA: the auth middleware answers htmx requests
   with `HX-Redirect: /login` (full-page client redirect), never a
   swapped login fragment.
3. If a partial action ever needs the whole page re-fetched (auth
   state changed), respond `HX-Refresh: true` — it's the full-reload
   escape hatch, and it's fine to use rarely and deliberately.
4. Auth state must be re-derived from the request context (session
   middleware), never from what a swapped fragment happens to contain.

## Component JS integration

1. `DOMContentLoaded` fires ONCE per document. Component controllers
   that must apply to swapped content bind via `htmx.onLoad(function
   (content) { … })` (scoped to the new content, not
   `document.querySelectorAll` over the whole page) — the canonical
   htmx re-init hook. A small dispatcher (components register init
   callables; the bundle runs them on initial load and re-runs them
   via `htmx.onLoad` for swapped content) is the standard shape.
2. htmx **executes `<script>` tags found in swapped content**. Therefore:
   partials must NOT contain scripts that re-declare global state,
   re-run app-bundle logic, or re-register global listeners.
   Enhancement scripts stay in the component registry bundle.
3. If JS adds DOM containing hx-* attributes dynamically, it must call
   `htmx.process(elt)` or the attributes are dead markup.
4. Floating widgets and chrome live outside the swap region — their
   state survives by construction; never move them inside it.

## History rules

1. **Every URL pushed with `hx-push-url` MUST render the full page on
   direct request** (copy-paste, new tab, history-cache miss). The
   Rule A path guarantees this — that's a reason it's non-optional.
2. `history:"reload"` (the pinned config) means back/forward always
   hits the server with a full browser load — no snapshots, no AJAX
   restore, sensitive data never enters a client cache. Do not opt
   into the v4 history-cache extension without an explicit decision.
3. `hx-history-elt` on the body is INERT under `history:"reload"` (no
   snapshots are ever taken) — it exists only so a future opt-in to
   the history cache works without a template change. Do not cite it
   as active behavior.
4. `HX-History-Restore-Request` exists in v4 but only matters on the
   AJAX-miss path that `history:"reload"` disables. Do not
   special-case it.

## Security

1. `html/template` auto-escaping is the first line of defense — keep
   every user/LLM-derived string escaped. Any raw `template.HTML`
   injection of untrusted content must scrub `hx-*`/`data-hx`
   attributes and `<script>` (htmx docs security rule: whitelist,
   never blacklist).
2. Wrap any raw, third-party-content region in `<div hx-disable>` so
   injected hx-* attributes can't issue requests. `hx-disable` cannot
   be un-disabled by nested content.
3. `htmx:validateUrl` event is the hook if same-host-only ever needs a
   narrower policy (today: `selfRequestsOnly` true is enough).
4. Never echo raw errors or provider payloads into swapped fragments —
   validation errors render through the error-summary component.

## Testing

1. Table-driven handler tests for the render contract, per page:
   - direct load → full document (doctype, sidebar, app.js tag, title)
   - `HX-Request: true` → the swap region only (no doctype, no
     sidebar, no script tags), 200, byte-identical to the full page's
     region — extract the region from the full render (anchor on the
     region's opening element AFTER `<body>`; inlined stylesheets can
     contain the selector as text and false-match document-wide
     searches); add every new navigation page to the table
   - error paths (unknown template, render failure)
   - when the partial carries an OOB fragment, the equivalence
     includes it — update the extraction and the renderer in the SAME
     commit whenever the swap-target semantics change
2. Pin hx-* attributes in render tests: chrome AND in-page nav links
   carry `hx-get`/`hx-target`/`hx-push-url` exactly; banned attributes
   (hx-* on auth forms, preference switchers, fragment links) are
   asserted ABSENT (ban-strings target rendered values, not labels).
   Attribute pins must `html.EscapeString` the hrefs — `html/template`
   escapes `&` to `&amp;` in attribute values.
3. Pinned copy in partials is unescaped before comparing
   (`html/template` entities: `&#43;` ≠ `+`).
4. Redirect tests assert BOTH status and headers: htmx redirect =
   200 + `HX-Redirect`; non-htmx = 303 + `Location`.
5. Every page gets a presence pin that the htmx bundle is in the
   shared layout (the every-page-presence family).
6. Behavior-neutral phases ship with a rendered-output diff vs the
   parent commit — an eyeball check is not evidence.

## Common traps (hallucination magnet — verify before trusting)

- **Header names are exact**: `HX-Request`, `HX-Redirect`, `HX-Refresh`,
  `HX-Location`, `HX-Push-Url`, `HX-Replace-Url`. There is no
  `HX-Redirect-Url`. In v4 the response headers are handled by
  DERIVING the key from the `HX-` prefix at runtime (`HX-Redirect` →
  `hx.redirect`) — so the literal header name does NOT appear in the
  minified dist; a grep miss is expected, not absence. v2's other
  response headers (`HX-Retarget`, `HX-Reswap`, `HX-Reselect`,
  `HX-Trigger*`) are NOT verified in v4 — treat them as gone until
  proven otherwise.
- `HX-Request` is the string `"true"` — compare with
  `r.Header.Get("HX-Request") == "true"`, not presence alone. v4
  still sends it on every request (verified in the dist), and adds
  `HX-Source` + `HX-Request-Type: full|partial` — available if ever
  needed.
- 302 + `HX-Redirect` does not work (3xx headers aren't processed).
- 4xx/5xx responses do NOT swap by default — "return the error page
  with status 400" silently renders nothing.
- `hx-select` on a server-rendered partial double-strips (or strips
  nothing when the selector doesn't exist in the fragment).
- **A partial that renders only the page's content block wipes the
  rest of the swap region** — the target usually contains the section
  nav, badges, and footer too; a content-only partial makes the
  section nav "disappear" on the first swap (2026-09-05, the
  nav-buttons regression). When the swap-target semantics change,
  the equivalence test and the partial renderer must change in the
  SAME commit — the partial is always the WHOLE region the target
  replaces.
- **A `load` trigger on a self-replacing fragment is a busy loop** —
  the response re-inserts the element with its own trigger, htmx
  processes the fresh element, and v4's `load` fires immediately at
  setup (not once per document): a self-refreshing span with
  `hx-trigger="load, every 30s"` issued ~165 requests/second
  (2026-09-05, a reported "app is slow" that was really a self-fetch
  loop). `load` belongs on one-shot lazy content — a placeholder whose
  response REPLACES it with content carrying no trigger — never on a
  fragment that re-inserts its own trigger. Pin loop-guard bans on
  both the layout element and the fragment endpoint.
- **Slow pages can split into an instant shell + a lazy fragment**
  (a fragment endpoint with raw named-template execution and a
  one-shot `load` loader): the user sees immediate feedback on both
  full loads and htmx navigation. Reconsider before using it — one
  global loading indicator may be the better default (2026-09-05: a
  per-page lazy split was reverted in favor of the global progress bar
  by operator decision). Use the split only for a page slow enough
  that the bar is not enough.
- Scripts in swapped content DO execute; `DOMContentLoaded` does NOT
  re-fire.
- `historyRestoreAsHxRequest` left at default breaks full-page history
  restore on cache misses (partial where a page belongs). — THE v2
  TRAP, RE-EXPRESSED IN v4: the config key is GONE; the same
  failure-mode is prevented by `history:"reload"` (full browser load
  on misses). Never remove that meta without an explicit decision.
- htmx 1.x advice (e.g., IE11 support, different defaults) is not
  htmx 2.x advice, and neither is htmx v4 advice — the migration
  guides exist for a reason. v4 changed event names (colon form) and
  dropped config keys; do not copy 2.x examples into v4 code.
- Titles: partials don't carry `<title>`, so per-page titles come from
  the full render only — don't fight `ignoreTitle` unless a partial
  legitimately swaps one in.
- Bundle-content pins that count literal tokens (`(function () {`)
  break when the token appears inside other code (`setTimeout(function
  () {`) — use named callbacks in shared bundles.

## Cheat sheet (verified against the pinned v4 dist, 2026-09-05)

Request headers: `HX-Request`, `HX-Source`, `HX-Request-Type`
(`full`/`partial`), `HX-Current-URL`, `HX-History-Restore-Request`
(present; inert under `history:"reload"`). Response headers: derived
from the `HX-` prefix at runtime — `HX-Redirect`, `HX-Refresh`,
`HX-Location` confirmed; others unverified (assume gone). Events
(colon form; the camelCase twins are NOT guaranteed in v4):
`htmx:before:request`, `htmx:after:request`, `htmx:finally:request`,
`htmx:before:swap`, `htmx:after:swap`, `htmx:before:settle`,
`htmx:after:settle`, `htmx:before:response`, `htmx:after:process` (the
v2 `htmx:load` equivalent — `htmx.onLoad()` wraps it), `htmx:after:init`,
`htmx:config:request`, `htmx:response:error`, `htmx:before:history`,
`htmx:after:history`, `htmx:before:view`, `htmx:after:view`,
`htmx:before:cleanup`, `htmx:after:cleanup`. Config keys (v4):
`history` ("reload" recommended — also the opt-in cache story), `logAll`,
`prefix`, `swapAfterOOB`… — v4's config surface is SMALL; verify any
key you need against the dist before pinning it. Config is set via
`<meta name="htmx-config" content='{...}'>` (merged) or JS.

Debugging: `htmx.logAll()` logs every event; `monitorEvents(elt)` in
the console shows what a DOM element actually fires; the unminified
`htmx.js` is ~2500 lines and has `issueAjaxRequest()`/
`handleAjaxResponse()` as the breakpoint targets.

Canonical sources: https://htmx.org/docs, https://htmx.org/reference,
https://templ.guide/server-side-rendering/htmx (Go integration),
`a-h/templ` `examples/counter` (full-page + hx-select pattern),
`jritsema/go-htmx-tailwind-example` (stdlib Go CRUD). The htmx QUIRKS
page (linked from the docs) covers edge-case behavior — read it before
relying on anything subtle.
