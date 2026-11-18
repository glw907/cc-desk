# Survey v1.7.0 has no sidebar rail, muted text, reading measure, or leading tokens

The package sheets define no `.rail*` classes, no `.muted` class, and no `--doc-measure` token.
The sidebar and the front page need all three, so `web/src/app.css` defines them from Survey's
tokens, under a comment that names the gap.

The same sheet carries a few literals Survey has no token for: the widths `--rail-width` (23rem),
`--doc-measure` (36.5rem), and `--page-width` (76rem), the 80rem breakpoint for the two-column
boards, the line heights 1.25, 1.35, 1.5 and 1.55, and the font weight 600. Survey's only weight token is
500. The polish adds two more gaps: the weight reset to normal on not-started words, and a 48px
section gap, which `app.css` writes as two 24px tokens because Survey's spacing tokens stop at 24px.

Fix by adding these to Survey and dropping the local copies from `app.css`. Found by
`page-front`.
