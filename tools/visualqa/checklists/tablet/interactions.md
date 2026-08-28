Tap targets: buttons, links, and nav items must be at least 44x44 CSS px (48dp) with visible spacing — touch and pointer coexist on tablets. Evidence in frame: estimate against the frame width — 44px is about 6% of 768px; smaller or touching targets are a FAIL, borderline is UNCERTAIN.
Hover never required: critical functions must be reachable by touch. Evidence in frame: hover-only content (tooltips, menus that only appear on hover) hidden in the frame is a FAIL.
Sticky overlap: fixed headers, sidebars, or bottom bars must not cover interactive content. Evidence in frame: a bar overlapping a button, link, or form field is a FAIL.
States: loading, empty, and error states must be visibly distinct and informative. Evidence in frame: a blank or content-less state with no cue is a FAIL.
