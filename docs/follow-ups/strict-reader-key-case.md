# The strict reader accepts keys in any case and duplicate keys

`encoding/json` v1 matches keys case-insensitively and lets the last duplicate key win, even with
`DisallowUnknownFields`. A release with `"Project"`, or with both `"PROJECT"` and `"project"`,
loads clean. `internal/state/read.go` `decodeStrict` is the place.

Fix when `encoding/json/v2` leaves `GOEXPERIMENT`, or by a token scan that rejects a duplicate or
mis-cased key. Found by the whole-branch review of `state-rules-files`.
