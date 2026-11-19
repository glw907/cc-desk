# Four small findings from the desk-open review

The whole-branch read of pass 3 accepted it and left four non-blocking findings.

- The question asked again after a failed check (`cmd/cc-desk/merge.go:146`) reads "Merge it
  anyway, send it back, or close it?". A "merge" answer asks the same question again while a check
  still fails, so the wording promises what the answer cannot do. Reword it to "Send it back, close
  it, or merge it once the checks pass?".
- `cmd/cc-desk/open.go:118` puts the skill file's path into a `Read(...)` allow rule without passing
  it through `desk.CheckCommandPath`, as the binary and state folder paths are. A path with a space
  or a glob character would write a rule that does not match. Only the hidden `--skill-file` flag
  can supply such a path.
- `cmd/cc-desk/open.go:139` writes `<state>/launch/desk-<id>.json` on each open, and nothing
  removes them.
- The "A wrong record" section of `skills/cc-desk/SKILL.md` points at `$XDG_STATE_HOME/cc-desk` for
  `rulings.jsonl`, but with `--state-dir` the folder is the one the system prompt's command names.
