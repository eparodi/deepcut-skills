# Mobile — all groups (28 checks). Concatenation of the group files.
1. Content fits the viewport width — no horizontal scroll or clipped content at 375-390px.
2. Nothing important is hidden under the notch, status bar, or home indicator (safe areas respected).
3. Key content and primary actions are reachable within a one-handed thumb zone (lower half of the screen).
4. The layout reflows correctly between portrait and landscape without breaking or losing content.
5. Body text is readable without pinch-zoom (effective font size at least 16px) with comfortable line height.
6. Text contrast meets WCAG AA (4.5:1 body, 3:1 large text) in both light and dark themes.
7. No truncation or ellipsis cuts off meaning; text wraps cleanly without overflowing its container.
8. Layout doesn't shift as web fonts or images load (no cumulative layout shift).
9. Primary navigation is reachable with one thumb (bottom tab bar or an easily reachable menu).
10. Back navigation is visible and works — no dead ends or trap pages.
11. Menu/drawer opens without hiding the current context and closes easily (scrim tap or close button).
12. The current location is clear (active nav state, breadcrumb, or page title).
13. Inputs are at least 44-48px tall and don't trigger page zoom on focus (font size at least 16px).
14. The keyboard type matches the field (email, number, phone, search).
15. Labels are visible and associated; the focus state is obvious; validation errors are readable without zooming.
16. Checkboxes, radios, and selects have touch-friendly hit areas — no tiny tap zones.
17. Images scale to their container and don't overflow, stretch, or distort.
18. Media keeps correct aspect ratios — no letterboxing, cropping, or distortion.
19. Videos and embeds are playable with touch controls and don't block content.
20. Lazy-loaded media doesn't cause jarring layout shift as it loads.
21. Tap targets are at least 44x44 CSS pixels (48dp) with adequate spacing between them.
22. Touches get immediate visual feedback (active states) — no dead taps.
23. Scrolling feels natural — no scroll traps, jank, or sticky elements covering content.
24. Loading, empty, and error states are visible and give a path forward.
25. Controls meet WCAG 2.2 Target Size (Minimum): 24x24 CSS pixels, or spaced so a 24px circle doesn't touch another target.
26. Text reflows at 400% zoom without horizontal scrolling or lost content.
27. Focus and screen-reader order match the visual order; interactive elements have accessible names.
28. Nothing relies on hover — every interaction works by touch (and has a keyboard alternative where relevant).
