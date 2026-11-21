# Sidebar headings: arc log

Geoff's dose, 2026-11-20: "a gentle sidebar session as part of the 1.1 brainstorm". His notes are
in `docs/follow-ups/sidebar-headings-and-title.md`. This reopens the "already-right" verdicts for
`.rail-heading` and `.rail-row` in `../page-polish/ledger.md`, "The rail, every page".

| Probe | Verdict | Why |
| --- | --- | --- |
| One probe page: today's rail beside candidates A, B, and C, light and dark, at `23rem`, on the `page` test state | Geoff picked A and sent back none of the shared changes | A puts the headings above their rows with the least added ink |

## What Geoff saw

Every candidate built the rail from daisyUI parts with Survey's tokens on top: a `navbar` holding
the title at `--text-step-13` with `site-benchmark`, "Rulings" as a `btn btn-ghost` in
`navbar-end`, the forecast as a `stat`, each section as a `menu` whose heading is a `menu-title`,
"Rough" as a `badge badge-soft badge-secondary badge-sm`, and each building round as one link in a
nested menu. The scale bar stays custom, since `progress` holds one value. No words changed.

- A, kept: headings in `--color-base-content`, weight 600, `--text-step-11`.
- B: A with a dashed `--site-contour` rule under each heading.
- C: headings in Spectral at `--text-step-12`.

## Measurements

Every text pair in A is above 4.5:1 in both themes. The lowest is the "Rough" badge in light, at
4.64. The stock `menu-title` color in daisyUI is 2.32:1 in light and 3.28:1 in dark on the rail's
ground, so A overrides it. That is a Survey gap to file when the page pass lands.

## Files

`candidate-a.patch` is the throwaway prototype's diff against `web/src` at `43c0669`. It is the
page pass's reference and was never applied. `before-*.png` and `candidate-a-*.png` are the rail
in both themes. The page pass keeps the names `web/tests/sidebar.spec.ts` asserts, or changes the
tests with the markup.
