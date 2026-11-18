# Page polish: audit ledger

Two read-only lenses graded the eight before shots against the four mock-up screens at
`d5c0435`'s `web/src`. Lens A (Opus, medium) graded typography, rhythm, and hierarchy by eye and
by the CSS. Lens B (Sonnet, medium) measured the built app at 1600 and 1200 wide in both themes
with `getComputedStyle` and `getBoundingClientRect`. Both found the work mostly right. Light and
dark graded the same on every knob, and 1200 matched 1600 except that the boards stack to one
column below `80rem`, as designed. The rulings below follow the plan's seven rules.

## Front page, `/`

| Knob | Lens | Verdict | Change | Impact | Ruling |
| --- | --- | --- | --- | --- | --- |
| Type scale, `.desk h2` step 12, `.board-head h3` step 11, `.ahead` and `.rounds` step 9 | A, B | already-right | | | |
| Leading of `.ahead` and `.next p` (1.5) | A | already-right | | | |
| Measure of `.desk .next` at `--doc-measure` | A | already-right | | | |
| Heading margin, h2 to text 12px | A, B | already-right | | | |
| Section gap, `.desk .next` to the boards is 24px | A, B | recommend | `margin-bottom: calc(var(--spacing-24px) * 2)` | medium | apply, pass (a). Two tokens multiplied, no new literal; the gap inside a section is 12px, so a break should be four times it |
| Weight of `.state` on `not_started` and `not_designed` rows (600, darker than "Building") | A | recommend | A proposed weight 400 and `--color-muted` | medium | apply the weight only, pass (a), as `.state:is(.not_started, .not_designed) { font-weight: normal }`. Declined the color: swapping to `--color-muted` is a color change past the dose and the conductor wants any such swap listed as a call |
| Baseline of `.s` and `.state` against the row link in `.round-line` and `.passes li` (about 3px apart) | B | recommend | `align-items: baseline` | low to medium | apply, pass (a), one line |
| `text-wrap` on `.desk h2` and `.board-head h3` | A | recommend | `text-wrap: balance` | low | apply, pass (b), one line |
| Right edge shared by state words, links, and bars | B | already-right | | | |
| `.figures` tabular numbers | A, B | already-right | | | |
| `.pass-links` | A | not-applicable (an inline run, not a column) | | | |
| Link and button targets under 24px (`.ahead` links 21.6px tall, `.branch-title` 17px) | B | already-right: each sits in a text run or a full row, the scan passes, and a 24px floor on `.branch-title` grows the rail row by 7px | | | declined |

## Round page, `/alpha/r2`, and pass page, `/alpha/r2/p13`

The two pages share every selector, so each line holds on both.

| Knob | Lens | Verdict | Change | Impact | Ruling |
| --- | --- | --- | --- | --- | --- |
| Type, `.doc-title` step 13, `.prose-doc` step 11, `.prose-doc h2` step 12, `.doc-meta` step 9 | A, B | already-right | | | |
| Measure of `.page-main`, 584px, 70.6 characters | A, B | already-right | | | |
| Gap above a section, last block to the dashed rule is 16px, equal to a paragraph gap | A, B | recommend | `.prose-doc h2 { margin-top: var(--spacing-24px) }` | medium | apply, pass (a). Beats `.prose-doc > * + *` on specificity, uses Survey's top spacing token |
| Gap below `.prose-doc h2` (16px, the same as above it) | A | recommend | `.prose-doc h2 + * { margin-top: var(--spacing-10px) }` | medium | apply, pass (a). Lens A wanted padding-top 20px too; declined, since the 24px margin already separates the sections |
| `line-height` of `.prose-doc h2` (inherits 1.5 against 1.25 on `.doc-title`) | A, B | recommend (A), below the bar (B: no h2 wraps in the fixture) | `line-height: 1.25`, an existing named literal | low | apply, pass (b), one line |
| Figures list and state line alignment | A, B | already-right | | | |
| `text-wrap` on `.doc-title` and `.prose-doc h2` | A | already-right (`balance` is set) | | | |
| `.doc-meta` to `.doc-title`, title to body | A | already-right | | | |
| Focus ring on each link family and the button (2px solid, offset 2px, both themes) | B | already-right | | | |
| The command's `code` looks like a fallback monospace | A | not a knob: Lens B measured `code` at 13.76px `--font-mono` | | | not changed |

## Rulings, `/rulings`

| Knob | Lens | Verdict | Change | Impact | Ruling |
| --- | --- | --- | --- | --- | --- |
| Table type, `.table th` 14px, `td` 14px | A, B | already-right | | | |
| The first column starts 16px inside the edge `.doc-title` shares | A | recommend | `.table :is(th, td):first-child { padding-inline-start: 0 }` | low | apply, pass (a), one line |
| Tabular figures on the time column | A | recommend | `font-variant-numeric: tabular-nums` on `.table td:first-child` | low | apply, pass (b), one line |
| Rhythm, title to table | A | already-right | | | |

## The rail, every page

| Knob | Lens | Verdict | Ruling |
| --- | --- | --- | --- |
| `.rail-heading`, `.rail-row`, `.session-time`, `.figures` | A, B | already-right | |
| Section gaps (24px), heading to row (4px), release to release (12px) | B | already-right | |

## AI-default

Lens A answered no. The nearest brushes are Survey's own identity: Spectral on a warm ground, with
a blue accent and an ochre secondary, dashed contour rules with 4px corners, two boards of like
content with no shadow, and no all-caps eyebrows, arrows, or pills.

## Decisions log

1. Pass (a), spacing, alignment, and hierarchy: the `.desk .next` gap, the `.prose-doc h2` gaps
   above and below, the `.state` weight on not-started rows, baseline alignment in `.round-line`
   and `.passes li`, and the first-column inset of `.table`.
2. Pass (b), type and token conformance: `text-wrap: balance` on `.desk h2` and `.board-head h3`,
   `line-height: 1.25` on `.prose-doc h2`, and tabular figures in the table's time column.
3. Pass (c), states, focus, and accessibility: no ruling reaches it. Lens B measured the focus
   ring, the targets, and the 24px floor and found them right, so pass (c) gets no commit.
4. No ruling swaps a color token, so the polish raises no call about color. The one color swap
   proposed (`--color-muted` on not-started words) is declined.
5. No ruling touches `.command-actions`, the terminal link, "Copy command", or the muted pass
   links, so none builds on a flagged call from pull requests 9 or 13.
6. Every ruling lands in `web/src/app.css`. No `.svelte` file changes, no word changes, and no
   class is renamed.
