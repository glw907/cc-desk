# Survey v1.7.0 has no sidebar rail, muted text, or reading measure

The package sheets define no `.rail*` classes, no `.muted` class, and no `--doc-measure` token. The
sidebar and the front page need all three, so `web/src/app.css` defines them from Survey's tokens
alone, under a comment that names the gap.

Fix by adding the three to Survey and dropping the local copies from `app.css`. Found by
`page-front`.
