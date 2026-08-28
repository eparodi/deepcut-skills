# Visual QA Checklist Library

Curated, source-backed QA checklists for the `visualqa` tool, one file
per **device × group**. Point `--checklist` at a single file for one
group, or at a **directory** to run every `.md` file in it (all groups).

```
checklists/
  mobile/   layout  typography  navigation  forms  media  interactions  accessibility
  tablet/   layout  typography  navigation  forms  media  interactions  accessibility
  desktop/  layout  typography  navigation  forms  media  interactions  accessibility
```

Sources for every rule: `docs/research/visual-qa-device-design-sources.md`
(Google web.dev, W3C WAI + WCAG 2.2, NN/g, Smashing, CSS-Tricks, MDN,
LukeW, A List Apart, Baymard).

## Usage

```bash
# one group, mobile (the primary design target)
go run ./tools/visualqa --url <url> --device mobile \
  --checklist @tools/visualqa/checklists/mobile/layout.md

# ALL groups for a device: pass the directory — every .md file inside
# is read, sorted by filename, and concatenated
go run ./tools/visualqa --url <url> --device mobile \
  --checklist @tools/visualqa/checklists/mobile/
```

## Rule format — be precise, or say you can't tell

Each rule states three things: the **exact bar** (numbers where they
exist), the **evidence to look for in the frame**, and when the frame
cannot show the evidence. The vision model is instructed: PASS only
when the frame clearly shows the criterion holding, FAIL only when it
clearly does not, and **UNCERTAIN when the evidence is not visible** —
never PASS on missing evidence. When you read a report, an UNCERTAIN
item means "this needs a human or a different check", not "the model
was unsure about a real pass".

Examples of the precision:
- "at least 44x44 CSS px (48dp)" for touch targets, "24x24 CSS px or
  spaced so a 24px circle doesn't touch another target" for WCAG 2.5.8
- "4.5:1 body / 3:1 large text (≥24px)" for contrast
- "60-80 characters (≈700-950px at 16px)" for line length
- explicit UNCERTAIN markers for anything a static screenshot cannot
  prove (keyboard behavior, zoom reflow, tab order, hover states)

## The groups

| Group | What it checks |
|---|---|
| `layout` | viewport fit, overflow, breakpoints, safe areas, orientation, density |
| `typography` | text size, contrast (AA), truncation, line length, layout shift |
| `navigation` | reachability, back affordance, overlays, location cue, keyboard/menus |
| `forms` | input size, labels, validation, hit areas, keyboard behavior |
| `media` | scaling, aspect ratios, playback, lazy-load alignment |
| `interactions` | target sizes, feedback, sticky overlap, states, hover |
| `accessibility` | WCAG 2.5.8 target size, reflow, order/names, keyboard op |

## Device notes

- **mobile** (375×812 @3, touch-only): thumb reach, safe areas,
  keyboard/zoom behavior — the most demanding profile.
- **tablet** (768×1024 @2, touch + hover): hybrid input, split-pane
  layouts, both orientations, hover never required.
- **desktop** (1440×900, pointer): hover, keyboard operation, window
  resize, density.

## Cost & limits

Each checklist item becomes one vision-model check (~384 tokens per
screenshot regardless of list size, so the cost driver is screenshots,
not checklist length). A full device run is 28 rules — under the
32-item response clamp the tool enforces. Keep custom checklists
focused: 8-12 rules is the sweet spot for model reliability.
