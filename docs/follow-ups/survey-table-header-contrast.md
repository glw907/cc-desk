# Survey v1.7.0's table header fails WCAG contrast on the page background

On the rulings page, the daisyUI `th` ink is `#777a7b` on `#f7f7f3`. That is a ratio of 4.02, under
the 4.5 that WCAG 2.2 AA asks for body text, and axe-core's `color-contrast` rule reports it on all
three headers (When, Question, Ruling).

`web/src/app.css` sets `.table th` to `var(--color-muted)` until Survey ships a fix. The change
darkens the ink within its hue, as Survey's rules ask.

Fix by darkening the table header ink in `tokens.json` so it clears 4.5 on `--color-base-100`, in
both themes, then dropping the local rule. Found by `page-surface`.
