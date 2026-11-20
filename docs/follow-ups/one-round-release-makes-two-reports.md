# A one-round release makes two reports

A release with one round falls due for two reports in turn: the round's, then the release's
(`ReportDue` in `internal/sessions/decide.go`, design line 378). The check starts the conductor
once for each, and Geoff reads two reports that cover the same passes.

Pass 6 (`desk-run`) checked the desk round's acceptance item 9, which says the conductor records
one report, on the conductor's first start. Its rehearsal then recorded and read the release
report too. The real cc-desk release has one round, so its run will meet the same pair.

For the 1.1 round: decide whether a one-round release needs both reports, or whether the round's
report also stands as the release's.
