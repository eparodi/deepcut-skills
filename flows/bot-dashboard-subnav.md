# Bot dashboard htmx — subnav regression checklist

Pinned regression for the 2026-09-05 operator bug: after an in-page
htmx navigation, the section sub-nav row must still be present (the
old partial render wiped it — the buttons "disappeared").

- The section sub-nav row (Dashboard/Stats/Equity/Capital/Dust on the
  summary section) is visible above the page content and lists all its
  buttons — none are missing after the navigation.
- The active sub-nav entry is marked current (class "here" /
  aria-current) and matches the page that was just opened.
- The page content below the sub-nav belongs to the targeted page
  (its heading and tables), not to the previous page.
- In the HTML evidence, the rendered page contains exactly one
  `nav.subnav` element, and its links carry `hx-get`,
  `hx-target=".main-wrap"`, and `hx-push-url="true"` attributes.
- The content region does not contain nested layout chrome (no second
  topbar, sidebar, or bottom-nav inside the content area).
