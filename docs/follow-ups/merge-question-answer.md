# What an answer to a merge question does

The check asks Geoff a merge question (`merge-<pass>-<k>`) when a pull request lies outside its
design or a pass has used its two fix attempts, and it holds the pass while the question is open.
Nothing yet says what each answer makes the check do: merge anyway, send the pass back with Geoff's
words, or close the pass without a merge.

Decide the answer options for each reason and the transition each one records, then add them to the
`ask` decision and the answer verb. Until then Geoff clears a held pass by merging or closing the
pull request himself, which the check notices on its next run.
