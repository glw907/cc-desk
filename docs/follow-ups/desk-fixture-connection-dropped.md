# The desk fixture uses a stop reason the binary never writes

The desk fixture's check log (`e2e/testdata/desk/check-log.jsonl`) and `e2e/brief_test.go:41` both
carry the reason `connection_dropped`. No `cc-desk` verb writes it, so the page's table of check
reasons has no entry for it. The page reads it through its fallback for an unknown code.

`web/tests/words.spec.ts` skips this one code when it checks that every reason the fixtures hold has
a word. That skip is the only exception in the test.

Replace the reason in both files with one the binary writes, such as `sign_in_expired`, and drop
the skip in `words.spec.ts`. The Go brief test and the page's fixtures then agree.
