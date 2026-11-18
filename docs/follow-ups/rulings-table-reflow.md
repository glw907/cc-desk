# The rulings table has no scroll wrapper for a narrow screen

The table on `/rulings` sits directly in the reading column. A long ruling with no break point
could force sideways scrolling at 320px wide, which WCAG 2.2 reflow (1.4.10) asks a page to avoid.
No fixture ruling is that long today, so no test shows it.

Fix by wrapping the table in a `div` with the `overflow-x-auto` class in
`web/src/routes/rulings/+page.svelte`. It is a markup change, so the polish pass left it. Found by
the `daisyui-a11y-reviewer` read of `page-polish`.
