# Visual QA Checklist Library

Curated, source-backed QA checklists for the `visualqa` tool, one file
per **device × group**. Pass any of them with
`--checklist @tools/visualqa/checklists/<device>/<group>.md`, or run
everything for a device with `.../all.md`.

```
checklists/
  mobile/   layout  typography  navigation  forms  media  interactions  accessibility  all
  tablet/   layout  typography  navigation  forms  media  interactions  accessibility  all
  desktop/  layout  typography  navigation  forms  media  interactions  accessibility  all
```

Sources for every item: `docs/research/visual-qa-device-design-sources.md`
(Google web.dev, W3C WAI + WCAG 2.2, NN/g, Smashing, CSS-Tricks, MDN,
LukeW, A List Apart, Baymard).

## Usage

```bash
# one group, mobile (the primary design target)
go run ./tools/visualqa --url <url> --device mobile \
  --checklist @tools/visualqa/checklists/mobile/layout.md

# every group for a device ("run all")
go run ./tools/visualqa --url <url> --device mobile \
  --checklist @tools/visualqa/checklists/mobile/all.md
```

## The groups

| Group | What it checks |
|---|---|
| `layout` | viewport fit, overflow, breakpoints, safe areas, orientation |
| `typography` | readability, contrast, truncation, layout shift |
| `navigation` | reachability, wayfinding, back/undo, active state |
| `forms` | input size, keyboard type, labels, validation, focus |
| `media` | scaling, aspect ratios, playback, CLS |
| `interactions` | touch/click targets, feedback, scrolling, states |
| `accessibility` | WCAG: target size 2.5.8, reflow, focus order, keyboard |

## Device notes

- **mobile** (375×812 @3, touch-only): thumb reach, safe areas,
  keyboard/zoom behavior — the most demanding profile.
- **tablet** (768×1024 @2, touch + hover): hybrid input, split-pane
  layouts, both orientations.
- **desktop** (1440×900, pointer): hover, keyboard operation, window
  resize, density.

## Cost & limits

Each checklist item becomes one vision-model check (~384 tokens per
screenshot regardless of list size, so the cost driver is screenshots,
not checklist length). A device `all.md` is ≤ 28 items — under the
32-item response clamp the tool enforces. Keep custom checklists
focused: 8–12 items is the sweet spot for model reliability.
