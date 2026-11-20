# Weigh the hidden row-list flag if the desk's registration is classed auth-data

The whole-branch review of pass 5 (sessions-timer) left six gaps in `scripts/live-check.sh`.
Pass 6 (`desk-run`) closed the first five. This one is left.

The hidden `--claude-command` flag lets a caller supply a row list. The ancestor-terminal check
is the real barrier. Weigh it again if `desk register` is ever classed as `auth-data`.
