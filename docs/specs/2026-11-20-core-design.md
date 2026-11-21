# cc-desk: one outcome for every line of the core configuration

This is the first of three specs Geoff agreed to on 2026-11-20. It covers the configuration
every Claude Code session on the workstation reads: the global `CLAUDE.md`, the permission
rules, the auto-mode entries, the hooks, and the three path rules. Each line, rule, entry, and
hook gets one outcome. The work lands as one pull request in the dotfiles repo
(`glw907/workstation`, checked out at `~/.dotfiles`), and nothing changes until Geoff approves
that pull request.

Geoff is asked to approve two things now: this design, and the exact text of the new global
file. He approves the pull request later, when the work is built and reviewed.

Two files sit beside this one:

- [2026-11-20-core-claude-md.txt](2026-11-20-core-claude-md.txt) is the full text of the new
  global `CLAUDE.md`. Geoff approves that exact text.
- [2026-11-20-core-fates.tsv](2026-11-20-core-fates.tsv) lists all 299 existing items and the 8
  new deny rules, one row each, with the outcome and the reason.

The second spec replaces the development process. What Geoff decided about it is recorded in
`~/Projects/initiatives/claude-config-alignment/specs/2026-11-20-process-decisions.md`.

## What is wrong today

The global file tells every session what to do, and the settings decide what a session may
run. Both grew one ruling at a time. Measured on 2026-11-20:

- **Rules that can't bend are written as text.** The file says never to run `git add -A`.
  Sessions ran it 181 times in 60 days. In 115 of those runs the command named paths, and in
  66 it staged a whole tree. Anthropic's
  guidance is direct: "If something has to happen every time, make it a hook, a deny rule, or a
  managed setting. `CLAUDE.md` is for guidelines, and Claude is allowed to decide a guideline
  doesn't apply."
- **The file is at its cap.** It is 120 lines, and the dotfiles gate fails at line 121. It carries 19 date
  markers, 15 bolded phrases, three rules stated twice, and three paths that don't resolve as
  written.
- **Old grants are still live.** Nine auto-mode entries were written for passes that have
  closed: six for the setup-tooling pass, which closed on 2026-11-14, and three for cc-desk
  rounds. Eight of the nine clear work that is done.
- **Some allow rules are dead or unsafe.** Twelve of the 144 allow rules match nothing, serve
  retired work, or permit something no session should do unreviewed. One prints a tunnel token
  into the transcript.
- **Two instructions contradict each other.** One auto-mode entry says a session never seeks
  approval in chat. The process text says approval now arrives in chat. The memory page warns
  that "if two instructions contradict each other, Claude may pick one arbitrarily."
- **The hooks have gaps.** Every hook but one is registered by bare name, so it is found only
  when the session's `PATH` includes `~/.local/bin`. `claude-tierguard` lets a command through when the hook
  fails to start. `claude-block-pass` refuses read-only searches that only mention a tool's
  name. `claude-tail-guard` is the hook that fires most (352 blocks since 2026-09-21) and it
  misses `gh run watch`, `make test`, and `go test`.

## What it is for

A session should read a short file of guidelines it can weigh, and meet a hook or a deny rule
wherever Geoff means "never". After this lands:

- Every line in the global file is true, says one thing once, and names a path that exists.
- Each rule that can't bend is enforced by a hook or a deny rule, with tests.
- No grant outlives the work it was written for.
- The process spec starts from a core it doesn't have to clean first.

## What it leaves out

- The development process. cairn and dubplate are running under it today, and they read
  the same global file and settings. The five stops, the words "pass" and "executor", the
  pointer to `pass-core`, the merge path through `approval-queue`, `claude-merge-guard`'s
  refusal text, and `claude-launch-guard` all keep meaning what they mean now.
  [Left for the process spec](#left-for-the-process-spec) lists each one.
- Skills, agents, and project `CLAUDE.md` files. Those belong to the third spec. Three
  lines leave the global file because a skill already says the same thing, so no skill is
  edited. No `~/.dotfiles/CLAUDE.md` is created.
- Project settings. Each repo keeps its own. None of them grants again what this spec
  removes.
- `~/.dotfiles/.claude/settings.local.json`. It holds three web-fetch allow rules for
  documentation sites and stays as it is.
- New kinds of control. No managed settings, no sandboxing, no new auto-mode `soft_deny` or
  `hard_deny` entries, and no `denyInsteadOfPrompting`. The process spec decides those.
- Three absolutes that stay guidelines. The global file says technical mail is read only
  through `tech-mail.sh`, that Claude never drives Geoff's Chromium profile, and that the
  Workflow tool runs only when he opts in. The docs say a rule that can't bend belongs in a hook or
  a deny rule. Each of these needs a list of the other routes to the same thing before a rule
  can be written, and nobody has made those lists. The third spec takes them up.
- Lines that fit a narrower home. The stow steps and the dotfiles gate matter in one repo,
  and the credential check is a procedure the `proton-pass` skill could hold. The memory
  page says such lines belong in a project file or a skill. They move in the third spec, which
  owns those files.
- `includeGitInstructions` and `attribution`. The memory page offers them as the fix when a
  commit rule in `CLAUDE.md` loses to Claude Code's built-in Git guidance. No session record
  shows that happening, so both keys stay unset.

## What would make it a failure

- A cairn or dubplate session stops mid-run because a rule it relies on changed meaning.
- A hook refuses ordinary work more often than it catches a real mistake.
- A rule ends up enforced in two places that disagree.
- The global file grows.

## Constraints

- The old process keeps working. Every line it depends on keeps its wording, apart from
  dropped date markers and three corrected paths.
- Changes reach running sessions. `~/.claude/CLAUDE.md`, `~/.claude/rules/`, and every hook
  script this spec changes are links into the `~/.dotfiles` checkout. Such a script changes for
  every running session the moment that checkout is pulled. The global file and the path rules
  reach a session the next time it loads them. `~/.claude/settings.json` is a generated file.
  It changes when `claude-tooling-sync apply` runs, and Claude Code then reloads permissions
  and hooks into running sessions with no restart.
- A hook that runs on every command must not break every command. `claude-merge-guard` runs
  on each Bash call with `onFailure: block`. A crash there blocks all shell work in all
  sessions.
- The global file has a 120-line cap that counts every line. The cap stays.
- `onFailure: block` needs Claude Code 2.1.295 or later. pavilion runs 2.1.329. Nobody has
  read the laptop's version.
- Output styles don't reach subagents, and the global file does. A line can't be cut on
  the grounds that the `writing-voice` output style says it.
- The `~/.dotfiles` checkout holds another session's uncommitted files
  (`claude/.claude/docs/bluefin-admin/google-workspace.md` and `bin/.local/bin/sheet-revision`).
  The build works in a worktree and never touches them.
- Custom code stays small. The spec adds checks to four existing hook scripts and writes no
  new one.

## Approaches

Three ways to do the review were on the table.

1. **Trim from the audit.** Run Anthropic's prompt audit on the global file and fix what it
   flags. It is quick. It only finds what the audit looks for, and the audit doesn't read the
   settings or the hooks, so stale grants and fail-open hooks would stay.
2. **Give every line an outcome.** Go through each line, rule, entry, and hook and decide: it
   stays as a guideline, becomes a hook or a deny rule, moves to a path rule, moves to a skill,
   is changed, or is removed. Nothing is skipped, and Geoff can overrule any row. It costs more
   review up front.
3. **Rewrite from a blank page.** Write the file Anthropic's guidance describes and add back
   only what is missed. It gives the cleanest result and the highest risk: cairn and dubplate
   depend on lines whose value isn't obvious until they're gone.

Geoff picked the second. The line-by-line review is assembled from documented guidance: the
docs say to prune the file, to move rules that can't bend into hooks or deny rules, and to
move rules for one kind of file into path rules. Going row by row is how this spec applies
them.

Nine agents did the review on 2026-11-20. One audited each of the four surfaces against
measured use, a second tried to refute each audit, and a ninth settled the 30 rows where the
two disagreed. The design session then made the final call on each. The outcomes below are
the result.

## What gets built

### The global file

The new text is [2026-11-20-core-claude-md.txt](2026-11-20-core-claude-md.txt), 119 lines. It
replaces the current file in one commit. The record groups the three path rules with the
global file, so it shows 77 rows there. Of the 74 statements in the current file:

| Outcome | Count | What |
| --- | --- | --- |
| Stays | 63 | 14 of them wait for the process spec with their wording unchanged |
| Removed | 5 | Two duplicates, and three lines the `writing-voice` skill or a path rule already covers |
| Moves to a skill | 3 | The initiatives line (`log-project`), the mail line (`runbook`), the tooling-layout line (`claude-infra`) |
| Becomes a hook | 2 | The staging rule and the force-push rule |
| Moves to a path rule | 1 | The line about checking a terminal UI |

The changes a reader will notice:

- Date markers and bold go. `git log` holds who ruled what and when. The docs advise
  against emphasis as a way to make a line stick.
- The staging rule changes. "Never `git add -A`" becomes "Stage files by name; a hook
  refuses whole-tree `git add` or `git commit -a` in a primary checkout." See the next section.
- The force-push line leaves the file. A hook enforces it.
- Two search rules become one, in the section about searching before debugging.
- The dotfiles section dissolves. Its lines join the sysadmin list, and its Git rules move
  to a section that covers every repo.
- Two sections turn into bullet lists, one rule to a bullet: the rules for every repo and
  the rules for agents. The memory page says bullets are easier to follow than dense
  paragraphs.
- Two lines say more exactly what they ask. "Ask, then run it" becomes "ask for the
  go-ahead, then run it", and the vendors line names what "the existing set" is.
- The `.env` rule lives once, under Secrets.
- Three paths are corrected: `~/.dotfiles/scripts/secrets/secret-set.sh`,
  `~/.claude/docs/test-accounts.md`, and `~/.claude/docs/unattended-work-guards.md`.
- The tail rule widens to "gates, test suites, and CI watches", to match its hook.
- Writing voice shrinks to three sentences. The skill holds the rest.

### Rules that become hooks

Both checks go into `claude-merge-guard`. Today that script strips quoted text and
here-documents, looks for `gh pr merge`, and starts no other program. The Git parsing and the
Git subprocesses below are new code in it. Its tests live in
`tests/test_claude_merge_guard.py` and run the hook as a subprocess. The hook, its tests, and
the two text changes land in one commit, so the file never claims a hook that isn't there.

**Which directory a check looks at.** The hook takes the first of these it finds: the
directory in `git -C <dir>`, a literal `cd <dir>` earlier in the same command, and the `cwd`
field of the hook's input. When the directory is a shell variable or anything else the hook
can't read from the text, it allows the command. The same directory decides the current
branch for a push with no ref argument. Every test passes a `cwd`.

**Force push.** The hook refuses a live `git push` (including `git -C <dir> push`) when both
of these hold:

- It forces: `--force`, `-f` or a short-flag cluster containing `f`, `--force-with-lease`,
  `--force-if-includes`, or a ref argument that starts with `+`.
- Its destination is `main` or `master`: named in the ref argument, reached through `--all` or
  `--mirror`, or reached through `HEAD`, `@`, or no ref argument when the current branch is
  `main` or `master`.

A forced push to any other branch passes. All seven forced pushes in the last 60 days went to
feature branches with `--force-with-lease`, so this check would have refused none of them.

**Whole-tree staging.** The hook refuses, in a primary checkout:

- a live `git add` whose only targets are whole-tree ones: `-A`, `--all`, `-u`, `--update`,
  or a bare `.`, `:/`, or `*`, and
- a live `git commit` with `-a`, `--all`, or a short-flag cluster containing `a`.

`git add -A docs/` passes, because it names a path. `git add -A .` is refused, and a bare `.`
counts as the whole tree wherever in the repo it runs. A primary checkout is one where
`git rev-parse --path-format=absolute --git-dir --git-common-dir` prints the same path twice.
The check doesn't apply in a worktree, in a repo with no commit yet, or under `~/.local/state`,
`~/.claude/jobs`, `$CLAUDE_JOB_DIR`, `/tmp`, or `~/.cache`. The refusal message tells the
session to stage by name. The tests need a repo the check applies to, and test folders sit
under `/tmp`, so the tests can replace the exempt list. They do it through the hook process's
environment, which a session's command text can't set.

A primary checkout is a repo's main folder, as opposed to a worktree made from it. The scope
comes from measured use. A session's worktree holds only that session's work, and about 142 of
the 181 runs happened in one. Eight more ran in a scratch folder, and the count didn't place
the last 22 or so. A primary checkout can hold another session's files, as `~/.dotfiles` does
today. Nine runs in 60 days happened in one.

**Merges by other spellings.** The hook already refuses `gh pr merge`. It also refuses
`gh -R <repo> pr merge` and `gh api -X PUT .../pulls/<n>/merge`, with the same message. The
message that points a session at `approval-queue` stays byte for byte, and a test compares it
with a copy of today's text.

**When Git itself fails.** Every Git subprocess the hook starts falls through on error: the
hook allows the command and lets Git report the problem. A command with neither `git` nor
`gh` as a word exits before any parsing, so `git -C <dir> push` is still caught.

### Permission rules

Counts move from 144 allow, 4 ask, and 15 deny to 132 allow, 2 ask, and 25 deny.

| Change | Rules | Why |
| --- | --- | --- |
| Remove (1) | `Bash(gh pr merge * --match-head-commit *)` | The merge hook refuses every `gh pr merge`, so the rule permits nothing |
| Remove (1) | `Bash(approval-queue merge *)` | `Bash(approval-queue *)` already covers it |
| Remove (3) | `Bash(podman build -t cairn-e2e *)`, the `podman run --rm --env-file` rule, `Bash(gh repo delete glw907/cairn-e2e-* --yes)` | The e2e environment they served was removed on 2026-11-01 |
| Remove (2) | Both `ssh ... geoff@mail.907.life *` rules | Never used. Sessions reach that server as `ward.907.life`, and those rules stay |
| Remove (1) | `Bash(scripts/ward-mail-delete.sh *)` | Written for the planned deletions of the mail move, which is finished |
| Remove (1) | `Bash(gh workflow run publish.yml -f promote=*)` | Never matched; cairn-cms's project settings carry the form that does |
| Remove (1) | `Bash(cloudflare/access.sh --token ward)` | Prints the tunnel token into the transcript |
| Remove (2) | Both `claude-mail.sh` rules | The script isn't built; the household-it pass that builds it adds an exact rule |
| Ask to deny (2) | `Bash(voipms-sms send *)`, `Bash(voipms-sms --did * send *)` | Geoff runs a send in his own terminal, where these rules don't apply. Line 11 of `claude/.claude/docs/bluefin-admin/phone-checks.md`, which says a session asks first, is reworded in the same commit |
| Add deny (3) | `Bash(hcloud firewall delete-rule *)`, `Bash(hcloud firewall replace-rules *)`, `Bash(hcloud api *)` | They reach what the existing `hcloud` deny rules refuse |
| Add deny (3) | `Bash(clasp delete *)`, `Bash(clasp undeploy *)`, `Bash(clasp delete-deployment *)` | Other spellings of the script delete that is already refused |
| Add deny (2) | `Bash(git * --output*)`, `Bash(git * --upload-pack*)` | They turn a read-only Git allow rule into a file write or a program run. The first covers both `--output=<file>` and `--output <file>` |

The other 132 allow rules, the two `wrangler r2 bucket delete` asks, and the 15 existing deny
rules stay. Fourteen of them wait for the process spec.

A deny rule matches the command text. The permissions page says such a rule "covers the
invocation Claude usually produces and isn't a security boundary around the program": a full
path to the binary or an `sh -c` wrapper gets past it. The ten rules here stop the ordinary
spelling. The rules that must hold against other spellings are the ones in the hook.

### Auto mode

The allow list goes from 20 entries to 11, and the environment list from 14 to 13.

- Removed, nine allow entries. The `gh pr merge` entry, which clears nothing because the
  merge hook refuses that command. Five of the six setup-tooling grants. The
  three grants for finished cc-desk rounds (`cc-desk-prove-owned`, `cc-desk-sessions-page`,
  `cc-desk-desk-round`). cc-desk's five-minute timer reads the release record
  `~/.local/state/cc-desk/releases/cc-desk.json`, which is still in place. Before the last
  grant goes, the build confirms the timer finds nothing in that record to launch, or moves
  the record to `done/`.
- Removed, one environment entry. Claude's send mailbox, whose script isn't built.
- Changed. The entry for documentation and skill edits loses one clause, "this never tells
  a session to seek approval in chat instead of the approval queue", which contradicts
  `pass-core`. The local-development entry keeps its D1 half and rewords the other half to "a
  process the session started, by the PID it recorded at start". The source-control entry
  drops "and merging PRs". The repository-visibility entry gets fresh lists from `gh`, and its
  closing sentence about world-readable history stays word for word.
- Kept on purpose. The sixth setup-tooling grant clears a sweep that is still scheduled for
  the start of Geoff's live-setup sitting. That sitting's close removes it.

`claude/.claude/tooling/automode-initiatives.json` ties one grant to its initiative, the
setup-tooling grant for configuring the clean room. That grant is removed, so its line leaves
the file too.

After the settings are applied, the session runs `claude auto-mode critique`, which the docs
describe as the review for custom entries, and then `claude auto-mode config`. Two entries
that stay are already marked for that review. The entry for moving a worktree's uncommitted
edits onto the default branch had no measured use in 60 days. The entry for Git cleanup after
a merge was in place on 2026-11-02 when a matching command was refused anyway.

### Hook registrations and scripts

Eighteen registrations become sixteen. The two that go are both `claude-awake-touch`.

- Every hook command is registered by a path that starts with `~/`: `~/.local/bin/<name>` for
  the 15 registered by bare name, and `~/.claude/gather-dotfiles.sh` as it is now. A gate
  test checks that each registered path exists and is executable. This lands no later than the
  `claude-tierguard` change below.
- `claude-tierguard` gains `onFailure: block`, so it no longer lets a command through when
  it fails to start.
- `claude-secret-guard` checks names for `.env` and `.env.<x>`, apart from `.env.example`,
  `.env.sample`, `.env.template`, and `.env.dist`. It refuses a `git add` that names such a
  file. It refuses a `git commit` when a staged name matches, when the commit names such a
  path, or when `-a` is given and a tracked file of that name is modified. When a whole-tree
  add comes before the commit in the same command, it checks untracked names too. It finds
  the repo the way `claude-merge-guard` does, and it allows the command when Git fails. The
  script's existing override, `CLAUDE_SECRET_GUARD_ALLOW=1` in the hook's environment, covers
  this check as it covers the rest. No repo under `~/Projects` tracks a matching file today;
  the only tracked ones are two `.env.example` files. The hook still doesn't scan file
  contents at commit time. In `~/.dotfiles` the pre-commit hook already runs `gitleaks` on
  staged content, and a dotfiles pass scans the branch history at close. The script's
  docstring is corrected to say it scans "the command text".
- `claude-block-pass` treats `(`, `)`, `|`, `&`, and `;` inside a quoted span as plain
  text only when the command holds no word that runs a shell. Those words are the script's
  `WRAP` list plus `sh`, `bash`, `zsh`, or `dash` with a `c` flag, `eval`, `ssh`, `su -c`,
  `script -c`, `find -exec`, and `claude-pass run`. With any of them present the script scans
  as it does today, so a new wrapper it doesn't know fails toward refusing. The must-allow
  tests are the read-only searches the hook refused in session records since 2026-11-01. The
  must-deny tests are `bash -c 'pass-cli item list'`, `eval "pass-cli item list"`,
  `sh -c "pass-sitting x.jsonl"`, `'pass-cli' item list`, `p"ass"-cli item list`, and
  `bash -c "secret-tool lookup a b"`. `ls docs/*pass*` is refused by a different rule, the
  session-directory glob, and the builder traces it separately.
- `claude-tail-guard` also covers `gh run watch`, `scripts/gate/run-gate.sh`, `make check`,
  `make test`, and `go test`.
- `claude-memory-write-check` changes only its path. An `if` filter would save a shell start
  on each write, but the hooks reference says one `if` names one tool, and it doesn't say how
  a path filter treats the `Write` tool. The script already exits on its first test for a
  file outside a memory folder.
- `claude-awake-touch` runs `async` on its two remaining events. Its `SubagentStart` and
  `SubagentStop` registrations are removed.
- The `desk` script is removed, with `tests/desk.test.sh` and its leg in `scripts/check.sh`.
  Geoff stopped running the desk session on 2026-11-17. `stow` leaves a dangling link behind
  when a file leaves its package, so after the pull the session checks that
  `~/.local/bin/desk` points into the dotfiles package and removes it. The laptop keeps a
  dangling link, which `command -v` skips.

### Path rules

`~/.claude/rules/go.md` gains one sentence: before any claim or gate about a terminal UI,
invoke `standards:tui-visual-verify`, because only a screenshot of the real terminal is
evidence. `cairn-family-ui.md` and `vale.md` name the old process and wait for the process
spec.

### Other settings

`skillOverrides` sets `anthropic-skills:google-workspace` to `"off"`. The global file already
says Google Workspace goes through `gws` and `clasp`, never a connector. The skills page says
`skillOverrides` doesn't reach a plugin's skills. The settings enable no plugin of that name,
and the eleven `name-only` entries for `anthropic-skills:` names are in the settings today, so
the build checks the result instead of assuming it. If the skill still runs after the change,
the entry goes back to `name-only` and the pull request says so. Every other settings key
stays as it is.

## How it lands

cc-desk's spec work doesn't run under the pass process (Geoff, 2026-11-20). This spec is built
by the flow on Claude Code's best-practices page. A fresh session builds from this file in a
worktree of `~/.dotfiles` cut from `origin/main`. `/goal` holds the definition of done, which
is "What done looks like" read against the worktree. There is no plan. The launch prompt lives
in cc-desk's `docs/plans/` and names this file.

Two fresh subagents read the result before the pull request opens. The first checks the diff
against this spec: every requirement built, gaps that affect correctness reported, style
skipped. The second reads the four hook scripts and the deny rules and looks for a way around
each. The build session fixes what they find or lists it in the pull request.

`claude/.claude/settings.json` sits in a `.claude` directory, which Claude Code protects. In
auto mode every write there goes to the classifier, and no allow rule can approve it ahead of
time. The classifier refused such edits in dotfiles worktrees between 2026-11-01 and
2026-11-20. The standing entries clear only allow rules, skill text, and `CLAUDE.md` text.
Geoff picked one auto-mode entry that clears the build's edits, inside its worktree and for
the files this spec names (2026-11-20). He approves its text in his terminal. The entry reaches
the settings through a pull request of its own, which lands before the build starts. The
build's pull request deletes it, so the grant ends at the merge it served. The classifier still
decides each edit. If it refuses one with the entry in place, Geoff answers a permission
prompt for that edit. The counts in this spec are the counts after the build's merge.

The build's pull request body names this file and the cc-desk commit that holds the approved
global-file text. It lists what each review found and what was done about it, carries the
one-time check from done item 2, and lists steps 4 to 7 under "After the merge". It has no
"Before merge" heading, because `approval-queue land` refuses a pull request that has one.

1. The session Geoff is talking to cuts the worktree, makes the settings edits there under the
   entry, and commits them. A fresh background session then builds the rest in that worktree
   under `/goal` and runs `bash scripts/gate.sh` there.
2. The two reviews run. The build session opens the pull request, and the session Geoff is
   talking to reports it in his terminal.
3. Geoff approves the pull request by number in his terminal, and that session merges it with
   `approval-queue land`. `claude-merge-guard` refuses every other merge command from a
   session, and this spec keeps the guard and its message. He can merge on GitHub instead.
4. The session runs `claude-worktree-busy ~/.dotfiles`. When no other session is writing
   there, it pulls. The hook scripts are now live in every session, and the global file and
   the path rule in every session that starts from then on. The other session's two files aren't in this pull request, so the pull leaves them
   alone.
5. It removes the dangling `~/.local/bin/desk` link, then runs `claude-tooling-sync apply` and
   `claude-tooling-sync verify`. The permission rules, auto-mode entries, and hook
   registrations are now live. No script is added, so `stow -R bin` isn't needed.
6. It runs `claude auto-mode critique` and `claude auto-mode config` and posts both outputs on
   the pull request, with the allow and environment counts read at that moment. Anything the
   critique flags becomes a backlog item in the dotfiles repo.
7. It runs one `gitleaks git` scan.

Between steps 4 and 5 the new hook scripts run under the old registrations. That is safe,
because no script is renamed and the old registrations name the same files.

One week after step 5, a session reads `~/.local/state/claude-denials.jsonl` and every refusal
the two new checks made, found in the session records. It marks each refusal right or wrong.
A check with more wrong refusals than right ones gets a backlog item to narrow or remove it.
The session files this look as a dated backlog item in the dotfiles repo after step 7.

## When something goes wrong

| What happens | What the system does |
| --- | --- |
| A hook refuses a command a session needs | The refusal says what to run instead. The session reads it and carries on. |
| A hook check crashes | `claude-merge-guard` allows the command whenever a Git subprocess fails. Its tests must cover a missing repo, a detached `HEAD`, and a directory that isn't a repo. |
| `claude-merge-guard` has a bug that refuses every command | Every Bash call in every Claude session is refused. Geoff's own terminal runs no hooks, so reverting the merge there with `git -C ~/.dotfiles revert` restores the old script at once. A session can also repair the script with the Edit tool, which this hook doesn't cover. |
| A registered hook path doesn't exist | The gate test fails before the pull request opens, so a missing path never reaches settings. |
| A removed allow rule or grant turns out to be needed | The auto-mode classifier decides the command. A refusal lands in `claude-denials.jsonl`, and the one-week check reads it. |
| The new global file or settings mislead sessions | `git revert` the merge, pull, and run `claude-tooling-sync apply`. Sessions get the old settings and hook scripts at once, and the old file the next time they load it. |
| The global file passes 120 lines during the build | The budget hook reports it right after the write, and the gate fails until the file is back under the cap. The file lands as one rewrite, at 119 lines. |

The limits: the global file stays at or under 120 lines; hook timeouts stay at 5, 10, 15, or
30 seconds as registered today.

## What done looks like

Geoff, or any session, can run each of these after step 5. Each one can fail. Where an item
says a command is refused, the refusal carries the hook's message, and the probe is a form
that does no harm if the hook misses it.

1. `git -C ~/Projects/cc-desk show <commit>:docs/specs/2026-11-20-core-claude-md.txt | diff ~/.claude/CLAUDE.md -`
   prints nothing. `<commit>` is the cc-desk commit the pull request body names, which holds
   the text Geoff approved.
2. A gate test in the dotfiles repo checks that every path in backticks in the global file
   that starts with `~/` exists. The pull request also shows a one-time check that each bare
   command and skill the file names exists, and that `pass-core` still has the headings
   "Judgment calls", "Until cc-desk runs", and "Direct changes".
3. `jq '.permissions | (.allow|length), (.ask|length), (.deny|length)' ~/.claude/settings.json`
   prints 132, 2, and 25. None of the 12 removed rules is present. All 8 new deny rules and
   both `voipms-sms` rules are in `deny`. A session's probe of each of those ten rules is
   refused; each probe is a `--help` or `--dry-run` form.
4. `jq '.autoMode | (.allow|length), (.environment|length)' ~/.claude/settings.json` prints 11
   and 13. By content: none of the nine removed grants is present in the settings,
   `automode-initiatives.json` no longer names the clean-room grant, the `setup-tooling-sweep` grant is present, the phrases "this
   never tells a session to seek approval in chat" and "and merging PRs" are gone, and "by the
   PID it recorded at start" is present. The build's auto-mode entry is gone too.
5. Every hook `command` in `~/.claude/settings.json` starts with `~/`, and there are 16 of
   them. `jq` shows `onFailure: block` on `claude-tierguard`, `async` on both
   `claude-awake-touch` registrations, no `args` key on any hook, and no
   `claude-awake-touch` under `SubagentStart` or `SubagentStop`. The gate test for registered
   paths passes.
6. The hook's tests cover a forced push with no ref argument on `main`, `+main`, `-fu`,
   `--all`, `git -C <dir> push`, and `--force-with-lease origin main`, each refused, and a
   forced push to another branch, allowed. Live, a session's
   `git push --force --dry-run origin main` in a primary checkout is refused.
7. The hook's tests cover `git add -A`, `git add .`, `git add -A .`, `git commit -a`, and the
   `cd <dir> &&` and `git -C <dir>` forms in a primary checkout, each refused, and
   `git add -A docs/` and every form in a worktree, allowed. Live, a session's
   `git add -A --dry-run` in a primary checkout is refused, and the same command in a
   worktree of that repo goes through.
8. The hook's tests cover a commit with `.env`, `.env.local`, or `.env.production` staged, an
   add followed by a commit in one command, and `git commit -a` over a modified tracked
   `.env`, each refused, and the same commit with `.env.example`, allowed.
9. The existing cases in `tests/test_claude_merge_guard.py` and
   `tests/test_claude_launch_guard.py` pass with no edits to them. A new case compares the
   `approval-queue` refusal text with a copy of today's, character for character, and two more
   cover `gh -R <repo> pr merge` and the `gh api` merge call.
10. `bash scripts/gate.sh` and `claude-tooling-sync verify` both exit 0 in `~/.dotfiles`.
11. `test ! -e ~/.local/bin/desk && test ! -L ~/.local/bin/desk` exits 0.
12. The pull request carries the outputs of `claude auto-mode critique` and
    `claude auto-mode config`, both run after `claude-tooling-sync apply`.
13. In a session started after the apply, `/anthropic-skills:google-workspace` returns the
    `skillOverrides` error and the skill doesn't run.

## What is left to the builder

- How the new checks are written inside the four hook scripts, and how their tests are laid
  out.
- The exact wording of each new refusal message, as long as it names what to run instead.
- The harmless command used to probe each new deny rule.
- The exact text of the build's auto-mode entry, which Geoff reads in his terminal.
- The order of commits inside the pull request, within two rules: the hook and its text change
  share a commit, and the `~/` hook paths land no later than `onFailure: block` on
  `claude-tierguard`.
- The refreshed repository lists for the visibility entry, read from `gh` on the day.

## Left for the process spec

Each of these stays as it is today because the old process relies on it.

**In the global file** (14 statements and two path rules): the instruction to keep working and
the five stops; the pointer to `pass-core` and its adapters; the line that overrides the
approval queue; the line that keeps state in STATUS; both worktree lines, which say
"executor"; the minute-estimate line; the unattended-guards line; the compact instructions;
three of the four process-proportionality lines; and the backlog and branch clauses in
`cairn-family-ui.md` and `vale.md`. The process spec rewrites them in the new words and for
sessions that never wait on a question.

**In the permission rules** (14): the three `gh release create` rules, the four
`cairn-run-gate` rules, `Bash(approval-queue *)`, the four `ssh` rules for the two servers,
and the two `wrangler r2 bucket delete` asks. The process spec decides which steps a session
may never run and lists them for Geoff's sitting.

**In auto mode** (3): the entry for skill edits under an approved pass plan, the entry for
`approval-queue land` and `merge`, and the environment entry that names the queue's port.

**In the hooks and scripts** (8): `claude-merge-guard`'s refusal text and its rename to
`claude-git-guard`; `claude-launch-guard`'s checks; the skill caps in `claude-context-budget`;
`round-watch`; `pass-report`; `approval-queue`; `claude-worktree-busy`; and
`claude-approval-queue-reminder`, which lives in the approval-queue repo and is registered
nowhere.

Four more things the review turned up for it:

- `permissions.denyInsteadOfPrompting` is documented and fits sessions that must never wait.
- A list of steps that are refused and held for Geoff would be new. Nothing in the docs
  describes one, so the process spec labels it invented if it wants one.
- 523 commands in the session records contain `gh pr merge` although the hook refuses it.
  Nobody has checked why sessions keep trying.
- Each of the six setup-tooling grants names a later step of Geoff's that removes it, and
  five outlived their work. Removing a grant belongs in a checklist a session runs at close.

## Found in other repos

The review read past the core in places. Nothing here is changed by this spec. The pull
request body repeats the list so each owner sees it.

- `initiative.yaml` still says `proposed` for cairn-setup, self-hosting, and theme-garden.
- Two `curl` commands in household-it's `mail.md` and `servers.md` match neither `curl` allow
  rule.
- The `approval-packet` skill's description names a reminder hook that isn't registered.
- poplar's project settings carry `bypassPermissions` config that project settings can't set.
- Hooks named `claude-clock-stop` and `drafter-page-hook` appear in session records and in no
  settings file.
- A node process listens on `127.0.0.1:7931` and nobody has identified it.

## Calls Claude made

Geoff decided the three-spec order, the boundary of this one, and the approach. These calls are
the design session's. The first six change something Geoff sees or does.

1. **The staging rule is narrowed.** It was an absolute. It becomes a hook that refuses
   whole-tree staging in a primary checkout and leaves worktrees alone.
2. **Sending a text message is refused.** Today a session asks first. After this, Geoff sends
   it himself.
3. **Date markers leave the global file.** His rulings stay; the "(Geoff, 2026-…)" tags go.
4. **The Google Workspace connector skill is turned off** in Claude Code.
5. **One finished-pass grant is kept** until his live-setup sitting.
6. **The build gets one auto-mode entry.** Geoff picked it over answering permission prompts
   (2026-11-20). He reads its text in his terminal, and the build's pull request deletes it.
7. The 120-line cap stays, and the file lands as one rewrite.
8. The new checks go into `claude-merge-guard`, and no new hook script is written.
9. `claude-tail-guard` is widened and its text line is kept, where the review had suggested
   deleting the line.
10. `claude-secret-guard` checks file names at commit time and leaves contents to `gitleaks`.
    dubplate tracks 28 allow-listed token fixtures a content scan would refuse.
11. The writing-voice pointer stays in the global file, because subagents don't get the output
    style.
12. The launch prompt lives in cc-desk's `docs/plans/`. There is no plan and no pass.
13. `~/.dotfiles/.claude/settings.local.json` and `claude-approval-queue-reminder` are left
    alone.

A reviewer read this design as the builder would and asked six questions. The answers are
calls of the design session too:

14. A check reads its directory from `git -C`, then a literal `cd`, then the hook input's
    `cwd`, and allows what it can't read.
15. Whole-tree staging means `-A`, `--all`, `-u`, `--update`, or a bare `.`, `:/`, or `*` with
    no named path.
16. The `.env` check covers `git add`, staged names, named paths, and `git commit -a`, and
    the script's existing override applies to it.
17. `claude-block-pass` relaxes its quote handling only when no shell-running word is present.
18. The dangling `desk` link is removed by hand on this host and left on the laptop.

A last read checked the settings mechanisms and the global-file text against Anthropic's
docs. It produced four more calls:

19. `claude-memory-write-check` gets no `if` filter, because the docs don't say how a path
    filter treats the `Write` tool.
20. Two sections of the global file become bullet lists, and two lines are reworded to say
    what they mean ("ask for the go-ahead", and the vendors line). The memory page asks for
    bullets over dense paragraphs and for instructions "concrete enough to verify". The
    process paragraphs keep their shape, because the old process reads them.
21. The Google Workspace skill override is checked after the apply and falls back to
    `name-only` if it doesn't take.
22. Three absolutes stay guidelines, the one-repo lines wait for a narrower home, and the two
    Git settings keys stay unset. [What it leaves out](#what-it-leaves-out) gives the reason for
    each.

## Where each mechanism comes from

Geoff asked that every mechanism carry a label. "Documented" means Anthropic's Claude Code
docs describe it. "Assembled" means it is put together from documented pieces. "Invented"
means it is ours.

| Mechanism | Label | Source |
| --- | --- | --- |
| A short global file of guidelines | Documented | Memory page: target under 200 lines; best-practices page |
| A rule that can't bend becomes a hook or a deny rule | Documented | Best-practices page |
| A `PreToolUse` hook that reads the command text | Documented | Permissions page: "To inspect the full command text with your own logic before it runs, use a PreToolUse hook." |
| Deny rules ahead of the classifier | Documented | Auto-mode configuration page: a deny rule "blocks before the classifier is consulted" |
| `onFailure: block` and `async` | Documented | Hooks reference |
| Path rules for one kind of file | Documented | Memory page |
| `skillOverrides` set to `"off"` | Documented | Skills page. Whether it reaches an `anthropic-skills:` name is checked at the build |
| `claude auto-mode critique` and `config` | Documented | Auto-mode configuration page |
| A fresh session builds from the spec under `/goal`, and a subagent reviews the diff against the spec | Documented | Best-practices page and the `/goal` page |
| A second review that attacks the hooks and the deny rules | Assembled | The subagent review is documented; its aim is ours |
| An `autoMode.allow` entry for one build | Assembled | The entry is documented on the permission-modes page; adding it in one pull request and deleting it in the next is ours |
| Geoff answers a permission prompt when the classifier refuses an edit the entry names | Documented | Permission-modes page: leave auto mode to run a step the classifier won't clear |
| Settings edits made first, by the session Geoff is talking to | Assembled | Both halves are documented; the split is ours |
| One outcome for every line | Assembled | The docs name the moves; the row-by-row pass is ours |
| Hook commands that start with `~/` | Assembled | The hooks guide gives a full path as the fix for "command not found". The hooks reference says a command with no `args` runs through `sh -c`, which expands `~/`. The SessionStart hook is registered this way today |
| The 120-line cap | Assembled | Stricter than the documented 200 |
| Refusing whole-tree staging only in a primary checkout | Invented | Scoped from 60 days of measured use; the rule itself is Geoff's |
| The one-week look at refusals | Invented | A plain check of whether the change worked |
