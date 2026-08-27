# Visual QA — Device Design Sources

> Sources for the per-device visual-QA checklists in
> `tools/visualqa/checklists/` (mobile / tablet / desktop).
> All URLs verified reachable 2026-08-26. Rules extracted per source;
> the "informs" tags map each source to the checklist groups it feeds
> (layout, typography, navigation, forms, media, interactions,
> accessibility).

## 1. Google web.dev — "Responsive web design basics"

**URL:** https://web.dev/responsive-web-design-basics/
**Author/Publisher:** Pete LePage & Rachel Andrew, Google
**Date:** 2019 (still the canonical reference)

**Key rules:**
- Always set the viewport meta tag (`width=device-width,
  initial-scale=1`); never disable zoom (`user-scalable=no` /
  `maximum-scale` cause accessibility failures).
- Users scroll vertically, not horizontally — content must fit the
  viewport (no horizontal scroll) or it's a poor experience.
- Images: `max-width: 100%` + explicit `width`/`height` attributes so
  the browser reserves space and nothing shifts on load (CLS).
- Layouts: flexible grids (flexbox/grid/multicol, `fr` units), not
  fixed pixels.
- Media queries can test capability (`hover`, `pointer`, `any-hover`,
  `any-pointer`) — never force a touchscreen user to act like they
  have a mouse.
- Breakpoints come from content, not device classes: start small and
  expand until a breakpoint is needed; add minor breakpoints for
  spacing/type tweaks.
- Reading: cap text width (~70–80 characters per line).
- Don't hide content just because the screen is small.

**Informs:** layout, typography, media, navigation.

## 2. W3C WAI — "Mobile Accessibility at W3C"

**URL:** https://www.w3.org/WAI/mobile/
**Publisher:** W3C Web Accessibility Initiative
**Date:** updated 2025

**Key rules:**
- Mobile accessibility is covered by the existing WCAG standards — no
  separate mobile guidelines; WCAG 2.2 applies to phones and tablets.
- WCAG 2.1/2.2 added success criteria that address mobile directly
  (target size, spacing, reflow).
- WAI-ARIA applies to mobile web apps, not just desktop.

**Informs:** accessibility.

## 3. Smashing Magazine — "Finger-Friendly Design: Ideal Mobile Touchscreen Target Sizes"

**URL:** https://www.smashingmagazine.com/2012/02/finger-friendly-design-ideal-mobile-touchscreen-target-sizes/
**Author:** Anthony (UX Movement), Smashing Magazine
**Date:** 2012

**Key rules:**
- Platform minimums: Apple HIG 44×44 pt, Microsoft 34px (26px min),
  Nokia 1 cm² (~28px) — but these undershoot the average finger.
- Average index finger: 16–20 mm ≈ 45–57 CSS px; average thumb
  ≈ 72 px (MIT Touch Lab fingertip study).
- Small targets cause errors (Fitts's Law) and are worse for thumb
  users; error rates drop as target size increases (two cited studies).
- Tablets have the space for finger-sized targets "liberally" —
  mobile is where sizing discipline matters most.
- When you can't fit finger-sized targets, follow the platform
  minimums and keep navigation minimal.

**Informs:** interactions, navigation, forms.

## 4. CSS-Tricks — "A Complete Guide to CSS Media Queries"

**URL:** https://css-tricks.com/a-complete-guide-to-css-media-queries/
**Author:** Andrés Galante, CSS-Tricks
**Date:** 2020

**Key rules:**
- Media queries target viewport AND capabilities: `pointer: coarse`
  (finger) vs `fine` (mouse), `hover`, `any-hover`, `any-pointer`.
- Use `prefers-reduced-motion`, `prefers-contrast`,
  `prefers-color-scheme` for accessibility-aware theming.
- Mobile-first = `min-width` queries; breakpoints by content, not
  devices (24,000+ Android viewport variants).
- Modern CSS (grid `auto-fill`/`minmax`, `clamp()`) reduces the need
  for width queries entirely.
- Bigger touch targets when the pointer is `coarse` / `hover: none`.

**Informs:** layout, interactions, accessibility.

## 5. LukeW — "Mobile First"

**URL:** https://www.lukew.com/ff/entry.asp?933
**Author:** Luke Wroblewski
**Date:** 2009

**Key rules:**
- Design mobile first because it forces focus: only the most important
  data and actions fit a 320×480 screen — you must prioritize.
- The result is an experience focused on key tasks without "interface
  debris"; that discipline improves the desktop experience too.
- Mobile platforms extend capabilities (location, touch, orientation)
  — design for them rather than degrading the desktop site.

**Informs:** navigation, layout (prioritization), interactions.

## 6. MDN — "Responsive web design" (CSS layout learning module)

**URL:** https://developer.mozilla.org/en-US/docs/Learn/CSS/CSS_layout/Responsive_Design
**Publisher:** MDN Web Docs (Mozilla)
**Date:** updated Dec 2025

**Key rules:**
- RWD = fluid grids + flexible images + media queries; HTML is fluid
  by default.
- Always include the viewport meta tag — mobile browsers otherwise
  lie about width (default 980px) and break narrow layouts.
- Media queries by viewport width AND capability (`pointer`,
  `any-pointer`); breakpoints with relative units, mobile-first.
- Never set text with `vw` alone — users lose the ability to zoom;
  use `calc()` with rem/vw.
- Media: `max-width: 100%` on `img`/`picture`/`video`; consider
  `srcset`/`sizes` so small devices don't download huge images.

**Informs:** layout, typography, media.

## 7. Nielsen Norman Group — "Mobile User Experience: Limitations and Strengths"

**URL:** https://www.nngroup.com/articles/mobile-ux/
**Author:** Raluca Budiu, NN/g
**Date:** 2015 (part of a long-running mobile-usability research project,
151 participants)

**Key rules:**
- Small screen → high content-to-chrome ratio: every added element has
  an opportunity cost; prioritize the essential.
- Mobile sessions are short and interruptible (~72s vs ~150s desktop)
  → design for interruptions: save state, gist before minutiae,
  simplify tasks.
- Single window → tasks must be self-sufficient in one app/site.
- Touchscreen → targets must be larger than mouse targets; typing is
  hard; undo is even more important on touch.
- Variable connectivity → light pages, fewer steps/page loads.
- Tablets are used differently from phones — different design
  constraints, different guidelines.

**Informs:** navigation, layout, forms, interactions (undo), typography.

## 8. A List Apart — "Responsive Web Design"

**URL:** https://alistapart.com/article/responsive-web-design/
**Author:** Ethan Marcotte
**Date:** 2010 (the article that defined RWD)

**Key rules:**
- The three technical ingredients: fluid grids, flexible images, media
  queries.
- Don't quarantine experiences per device ("an iPhone website") —
  treat devices as facets of one experience, progressively enhanced.
- Media queries let you "surgically correct" layout as it scales;
  increase link target areas on small screens (Fitts's Law).

**Informs:** layout, interactions.

## 9. W3C — "Understanding SC 2.5.8 Target Size (Minimum)" (WCAG 2.2)

**URL:** https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html
**Publisher:** W3C Accessibility Guidelines Working Group
**Date:** updated 2026

**Key rules (the exact numbers):**
- Targets for pointer inputs must be at least **24×24 CSS pixels**,
  unless spaced so that a 24px-diameter circle centered on each
  undersized target doesn't intersect another target.
- Exceptions: inline text links, user-agent controls, equivalent
  controls, essential (e.g., map pins).
- Independent of zoom; applies to touch, mouse, and pen — helps
  tremor/mobility users and touchscreens alike.
- Best practice: aim for the stricter 2.5.5 Target Size (Enhanced)
  (44×44 CSS px) for important controls — which matches Apple's
  44pt and Material's 48dp guidance.

**Informs:** accessibility, interactions, forms.

## 10. Baymard Institute — "Mobile E-Commerce UX" (research program)

**URL:** https://www.baymard.com/research/mcommerce-usability
**Publisher:** Baymard Institute (20,000+ hours of mobile usability
testing; 138 mobile sites benchmarked)
**Date:** ongoing

**Key rules:**
- Mobile traffic is often ~50% of e-commerce traffic, yet mobile
  conversion lags desktop badly — the mobile platform is genuinely
  hard: the lack of screen real estate causes a dramatic loss of page
  overview.
- Weakest mobile UX areas across the industry: main navigation,
  search autocomplete, forms, and site-wide features/elements.
- Forms research (their 400+ mobile guidelines) emphasizes: correct
  keyboard types, field labeling/placement, input optimization
  (smart defaults, autodetection), and readable validation errors.

**Informs:** forms, navigation, layout.

---

## Source → checklist-group mapping

| Group | Primary sources |
|---|---|
| layout | 1 (viewport/overflow/breakpoints), 4 (capability queries), 6 (fluid grids), 8 (RWD ingredients) |
| typography | 1 (line length), 6 (no vw-only text), 7 (content-to-chrome) |
| navigation | 1 (don't hide content), 5 (prioritize), 7 (short sessions), 10 (weakest area) |
| forms | 3 (finger targets), 7 (typing is hard), 10 (forms weakest area) |
| media | 1 (image sizing/CLS), 6 (responsive images) |
| interactions | 3 (touch targets/Fitts), 4 (coarse pointer), 5 (thumb), 7 (undo), 9 (WCAG 2.5.8) |
| accessibility | 2 (WCAG applies to mobile), 4 (prefers-*), 9 (target-size numbers) |
