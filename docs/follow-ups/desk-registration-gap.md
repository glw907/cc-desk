# A desk session has no verb to register itself

Closed by pass 5 (sessions-timer). `cc-desk desk register` registers an interactive terminal session as a desk and refuses every other caller (`TestDeskRegister`, six cases). The live run never ran it against `~/.local/state/cc-desk`, as the plan's answer 66 requires.

`prompt yes`, `prompt no`, and `rules add` refuse any caller that is not a registered desk with a
terminal, and the e2e tests register one by writing `desks.json`. No verb does that, so a real
desk session can't answer a prompt yet. Round plan answer 17 has the conductor, which runs in
Geoff's terminal, register as a desk for pass 5's live run, and that needs this verb or a hand edit
of `desks.json`.

Decide before pass 5 whether `cc-desk desk register` belongs in the command, or whether the pass 5
script writes the entry through `desk.Entry`.
