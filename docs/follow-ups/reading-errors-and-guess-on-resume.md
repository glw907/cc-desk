# A cached failed reading holds starts for five minutes, and a guess on a resume is dropped

Two small edges in the start rules. Neither blocks the pass.

`usage.Cached` in `internal/usage/usage.go` returns a fresh reading even when it carries an error.
One failed `claude` call therefore keeps every start waiting on "the usage figure can't be read"
for up to five minutes, with no new attempt. Skipping a cached reading whose `Error` is set would
retry at once. The cost is one more `claude` call per verb while the reader stays broken.

`pass start --guess Q` on a `stopped` pass takes the resume branch in `internal/state/events.go`
and ignores the guess without saying so. A one-line refusal or a printed note would tell the
caller the guess was not recorded.

Found by the whole-branch review of pull request 5. Geoff's call on the first, since it trades
retry cost against waiting time. The second is a wording choice the next pass can make.
