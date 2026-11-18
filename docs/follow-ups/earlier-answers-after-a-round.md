# Checking earlier answers after a round finishes

The main design says that when a round finishes, the release's conductor checks each answer
Geoff already gave for a later round against what the finished round changed. A question whose
ground changed is asked again, with his earlier answer shown
(`docs/specs/2026-11-17-cc-desk-design.md`, "Stale answers").

Nothing does this yet. No verb reopens an answered question, and the conductor's prompt doesn't
ask for the check. The Desk and skills round left it out: its releases have one round each, so
no answer waits for a later round.

The Move round is the first with a release of several rounds, cairn's. Its design should say
which verb reopens a question, what the desk reads out when it asks again, and where the
conductor's prompt asks for the check.
