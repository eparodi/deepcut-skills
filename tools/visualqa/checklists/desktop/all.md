# Desktop — all groups (28 checks). Concatenation of the group files.
1. The layout makes good use of the width — multi-column layout with a sensible content max-width; no stretched full-width text lines.
2. Common window sizes (1440, 1280, 1024) render without horizontal scroll or broken grids.
3. Resizing the window reflows content across breakpoints instead of squeezing or overlapping it.
4. Density is appropriate for a pointer device — content is not oversized to mobile proportions.
5. Body text is readable at desktop distance; reading content caps line length (~60-80 characters).
6. Text contrast meets WCAG AA (4.5:1 body, 3:1 large text) in both light and dark themes.
7. No truncation or ellipsis cuts off meaning in dense UI (tables, navigation, cards).
8. Layout doesn't shift as web fonts or images load (no cumulative layout shift).
9. Navigation fits the desktop form factor — top nav or sidebar with working hover states; the current location is clear (active state, breadcrumbs).
10. Keyboard navigation works: Tab order is sensible, focus is visible, Enter and Escape behave.
11. Dropdown menus work with mouse hover AND keyboard focus — no hover-only traps.
12. Deep pages have breadcrumbs or back affordances.
13. Inputs have visible labels, focus rings, and clear error placement.
14. Tab order through forms is logical; browser autofill works.
15. Validation errors are clear, inline, and announced accessibly.
16. Buttons and dropdowns have adequate hit areas (at least 32px) with spacing between them.
17. Images and media scale with their columns — no overflow, stretch, or distortion.
18. Media keeps correct aspect ratios; hover previews (if any) work.
19. Video controls are visible and functional.
20. Lazy-loaded media doesn't cause jarring layout shift as it loads.
21. Hover states give feedback on interactive elements (links, buttons, table rows).
22. Click targets are at least 32x32 CSS pixels with spacing — no overlapping interactive areas.
23. Scroll behavior is sane; sticky headers don't cover content; no scroll traps.
24. Loading, empty, and error states are visible and give a path forward.
25. The page is fully operable by keyboard: Tab, Enter, Space, Escape, and arrows work; focus is always visible.
26. Contrast meets WCAG AA in both themes; focus order matches the visual order.
27. Controls meet WCAG 2.2 Target Size (Minimum): 24x24 CSS pixels, or spaced so a 24px circle doesn't touch another target.
28. Screen-reader basics: landmarks, heading structure, accessible names; dynamic content is announced.
