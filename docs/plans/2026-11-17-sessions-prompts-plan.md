# Prompts: pass plan

Initiative: cc-desk
Round: sessions
Pass: sessions-prompts

Class `runner`. Planned 60 minutes; the path below is 110, under the 120-minute park line
(round plan answer 14). Design: `docs/specs/2026-11-17-sessions-design.md` (approved
2026-11-17), sections "Permission prompts" and "Follow-ups through the command", and the third
row of "The passes". Round plan: `docs/plans/2026-11-17-sessions-round.md`. Pass 2's plan,
`docs/plans/2026-11-17-sessions-launch-plan.md`, holds the launch this pass relaunches through.
Writing a person reads (the PR body, STATUS lines, follow-ups) follows
`~/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/sessions-prompts`, branch
`sessions-prompts`, cut from `origin/sessions-launch` at `e604def` (pull request 10, open, which
carries pull request 8). PR base: `main`. Push: `glw907/cc-desk`. Before the PR is ready, merge
`origin/sessions-launch` in again if it moved, and `origin/main` if it moved. On 2026-11-18,
`origin/sessions-launch` was still `e604def` and `origin/main` was `87c5eb6`, ahead by STATUS
commits only.

Scope: the hook verb records each prompt; `prompt yes` and `prompt no` answer it from a desk
session under the ruling guard, end the session, and relaunch it uncounted; `rules add` and
`rules remove` change a project's allow rules under the same guard; `followup` files a follow-up
in the session's worktree; and each answered prompt is filed as a follow-up. Merging is pass 4.
This pass edits none of `scripts/gate.sh`, `.github/workflows/`, or `web/`, which are the Page
round's.

## Pre-flight

No step in this pass needs Geoff, and no step runs the real `claude` or `systemd-run`. Every
write is inside the worktree or a test's `t.TempDir()`.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read 2026-11-18, 130 rows,
  2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z, all inside the 30-day window. By tool: Bash
  112, Edit 14, Write 2, SendMessage 1, ScheduleWakeup 1. Rows naming this pass's commands:
  `git push` 4 (to main, a branch delete, or chained; none refused a pass branch); `git commit`
  4 (chained, `--no-verify`, or an outage); `go test` 1 (chained after a `kill`); a settings
  file named in the target 10, one of them new: on 2026-11-18T03:06:17Z a Write to
  `spikes/4-permissions/settings.json` was refused as Self-Modification. `gofmt`, `gate.sh`,
  `gh pr create`, and `gh pr ready`: 0. The planner prompt
  (`docs/plans/2026-11-17-sessions-3-planner-prompt.txt:23`) allows the live hook probe only if
  the round plan's auto-mode entry is in place for it. The entry `cc-desk-sessions-page`
  (`~/.claude/settings.json:426`) grants launches only to the timer and the conductor, through
  the binary, of `cc-desk-scratch-*` sessions. It doesn't cover this executor, so this pass runs
  no probe. The fixture is the documented input, and pass 5 records a live one.
- Reviewed: one `standards:plan-reviewer` read of this plan (Opus, medium), 2026-11-18. It
  returned 4 high, 9 medium, and 6 low findings, all folded; the list is the last section,
  "Review findings and their folds".
- Evidence: the mechanism table and the recorded searches below. Each acceptance check was run
  on `e604def` before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| A write to a Claude settings file | No file this pass writes, fixture or test, is named `settings.json` or sits under `~/.claude`. The binary's settings files go under the test's state folder |
| `kill` or `pkill` around `go test` | Stop a background command with `TaskStop` |
| `git commit --no-verify`, `git stash` | Commit by path with hooks on. No stash |
| A chained `git push` or `git fetch` | Fetch, then push `sessions-prompts`, each alone |
| `gh pr merge` | The pass merges nothing |
| A classifier outage | Wait 60 seconds and run it once more, then flag the task and go on |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| The verbs today | `go build -o $T/cc-desk ./cmd/cc-desk` on `e604def`. `cc-desk --help` lists `check`, `completion`, `free`, `help`, `pass`, `prompt`, `release`, `round`, `ruling`, `session`, `snapshot`, `usage`. `prompt yes`: `unknown command "yes" for "cc-desk prompt"`, exit 1. `prompt no`: the same for `no`. `followup --help`: `unknown command "followup" for "cc-desk"`, exit 1. `rules --help`: the same for `rules` |
| The hook stub | `cmd/cc-desk/prompt.go`: reads stdin to the end, prints nothing, exits 0. Run with the documented input below and `--state-dir $T/st`: exit 0, no output, and `ls -A $T/st` empty |
| The hook's input | [Hooks reference](https://code.claude.com/docs/en/hooks), "PermissionRequest input", quoted: `{"session_id":"abc123","transcript_path":"/Users/.../.claude/projects/.../00893aaf-19fa-41d2-8238-13269b9b3ca0.jsonl","cwd":"/Users/...","permission_mode":"default","hook_event_name":"PermissionRequest","tool_name":"Bash","tool_input":{"command":"rm -rf node_modules","description":"Remove node_modules directory"},"permission_suggestions":[{"type":"addRules","rules":[{"toolName":"Bash","ruleContent":"rm -rf node_modules"}],"behavior":"allow","destination":"localSettings"}]}`. The page also says `permission_suggestions` is optional, a rule's `ruleContent` is optional, MCP tools add `mcp_server`, and common fields include `prompt_id`, `scratchpad_dir`, and `effort`. So the hook decodes leniently: unknown fields are ignored |
| The hook's exit | Same page: "Exit code 0 with no output means the hook has no decision to report, so the tool call continues through the normal permission flow." "Exit code 2 isn't honored for this event." Exit 1 is a non-blocking error unless the hook sets `onFailure: "block"`, which denies. So the hook always exits 0 with empty stdout, and any failure goes to stderr |
| When the hook fires | Same page: in "sessions that can't show a prompt, such as background subagents in non-interactive mode, Claude Code still runs these hooks, and if no hook returns a decision, it denies the tool call". `docs/research/2026-11-17-prove-findings.md` proof 2: a `--bg` session shows a prompt, `waitingFor: "permission prompt"`. The spike's hook fired in a `--bg` session in `default` mode on `date -u +%s > probe.txt` (prove findings, lines 138 to 157); no raw input was saved |
| The hook under auto mode | Not run. The design launches every session with `--permission-mode auto`; whether an unallowed command under auto mode ever reaches the hook, or the classifier decides it first, is unprobed. Pass 5's live run meets it. This pass's tests feed the hook its input directly |
| The hook's `session_id` against the record | Not run. Whether the input's `session_id` equals the `sessionId` that `claude agents --json` gives is unprobed. The hook matches on `session_id` alone and keeps the raw input of a miss, so pass 5 can read it |
| The commit path a verb takes | `cmd/cc-desk/root.go:56-93` `f.commit`: "Every verb takes a usage reading first". With no `--usage-command`, the reading runs `claude -p /usage` with a 60 s timeout (`internal/usage/usage.go:26, 63-70`). The hook's command line carries no such flag (`internal/launch/settings.go:72`) and has a 30 s timeout. So the hook calls `state.Commit` directly with no reading |
| What a commit settles | `internal/state/commit.go:80-88`: after the event, `settle` runs unless `s.noSettle` is set. `checkrun.go:88-95` `RecordListFailure` sets it so "a waiting pass stays waiting". So the hook's event sets `noSettle` too |
| What a commit checks | `commit.go:80-114` applies the event and writes, with no second validation. `check.go:262-266` refuses at load a prompt with no time or no tool, and the plan-reviewer's probe showed one such prompt fails every verb after it (`releases/chk.json: pass "s01": prompt 1 needs "time" and "tool"`). So `RecordPrompt` refuses an empty tool or a zero time, and the hook treats a missing `tool_name` as a bad input |
| The ruling guard | `internal/state/ruling.go:39-53` `guard(c Caller)`: the caller's session id is a registered desk, the caller's process chain holds the desk's PID with the same start time, and that process has a terminal. The refusals at `ruling.go:16-18` start with `cannot record the ruling: `, and `e2e/ruling_test.go:245-247` and `internal/state/ruling_test.go:66-103,258` compare them. `ruling.go:105` uses the prefix in "already answered". `internal/desk/desk.go:132` `Caller()` reads `CLAUDE_CODE_SESSION_ID` and the process chain. So a background session's own id is not a desk and is refused at test 1 |
| No verb registers a desk | `ruling.go:123-124`: `RegisterDesk` "The Desk round calls it when a desk opens; no verb does." `e2e/ruling_test.go` `recordRuling` writes `desks.json` itself. So tests register a desk by writing that file, and pass 5's conductor has no verb to register with (Decisions) |
| Where a prompt lives | `internal/state/types.go:121-127` `Prompt{Time, Tool, Command, SuggestedRule, Answer}` on `Session.Prompts`. `state/launch.go:102` `RecordLaunch` carries `Prompts` over on a relaunch |
| Which records are live | `internal/state/check.go:412-414` `LiveSessions` visits every session record of each live release, "ended ones included". So this plan's "open record" means a session record with no `ended` time; whether its row is live is a list read |
| How a prompt stop reads | `internal/sessions/classify.go:49-51`: a row in state `blocked` with `waitingFor` `"permission prompt"` is labelled a permission prompt. `decide.go:297-300` `blockedRow` returns `res(Stop, "permission_prompt")` with no detail. `internal/state/checkrun.go:73-80` `applyRecorded` stops a pass through `StopPass`, or a conductor's or design writer's session record by `StopReason`, and sets `StopDetail` from the decision. `internal/snapshot/sessions.go` `todoFor(StopPermissionPrompt)` reads "answer it at the desk". Tests expect a detail of the form `Bash: rm x` (`checkrun_test.go:118`, `snapshot/sessions_test.go:29`) |
| Resuming a stopped pass | `internal/state/events.go:191-202` `StartPass` resumes a stopped pass, closes its stop interval, clears its stop, sets building, and checks no limits (`docs/follow-ups/resume-skips-the-limits.md`). `RecordLaunch` (`state/launch.go:85`) refuses unless the pass is building |
| A record the check left stopped | `internal/sessions/decide.go` `early`: a pass stopped for anything but usage gets `none`. So the check never relaunches an answered record; the answer verb does |
| The launch | `cmd/cc-desk/launcher.go:178` `Launch`: refuse on conflicts (skipping the record it relaunches, `:196`), reuse the worktree for a relaunch, `launch.Verbs`, then `WriteSettings(l.dir, l.binary, pl.name, rf.AllowRules, verbs)` from a fresh load, so a rule committed before the call is in the new file. `launcher.go:234-243`: a relaunch's count is `max(old.Counted(), 1)`, or one more when counted, and `p.Relaunch` is set on every relaunch. `launcher.go:165-170` `launchRequest{Kind, Cause, Counted, Replaces}` has no field for an answer note. `RecordLaunch` builds the new record from the old one without its stop (`launch.go:90-113`) |
| Stopping a session first | `cmd/cc-desk/carryout.go:47-95` `carry`: `sessions.List`, `runner.Stop` on a live row, a second list read, and "still live" gives "not relaunched". `docs/follow-ups/launch-review-notes.md` item 1: the real `claude stop` on a gone id may exit nonzero, so the answer lists first and stops only a live row. `launchOrStop` calls `failedRelaunch`, which stops the pass or the session record as `launch_failed` |
| The lock | `cmd/cc-desk/session.go:66-70`: `session start` takes `check.lock` (`state.CheckLockFile`, `state.LockWait` 10 s, `internal/state/lock.go:14`) for its whole run. A check holds it through its second-read wait, 30 s by default (`launcher.go:36`). So the answer waits up to 60 s for it |
| The settings and the verbs | `internal/launch/settings.go:42-55` `Verbs`: exact report commands from `binary + " --state-dir " + stateDir + " "`. `:61-87` `WriteSettings` wraps each verb as `Bash(<verb>)`. `internal/launch/templates/pass.txt` lists `.Verbs` under "Report a stop with one of these commands, and nothing else"; `design.txt` and `conductor.txt` list them the same way. So a follow-up command can't sit in `.Verbs` |
| The relaunch wording | `internal/launch/templates/relaunch.txt`: "This is relaunch {{.Relaunch.Count}} of 2, after {{.Relaunch.Cause}}. ..." `internal/launch/launch_test.go:108` and `e2e/recovery_test.go:169` pin "This is relaunch n of 2". `launch_test.go:99` fails a first-launch prompt holding the word "relaunch". So the answer relaunch gets a template of its own, `relaunch.txt` stays, and the follow-up line never says "relaunch" |
| The permission rule form | [Configure permissions](https://code.claude.com/docs/en/permissions): `Bash(npm test)` is an exact rule. "Every subcommand must be allowed for a compound command to run without a prompt," with `&&`, `\|\|`, `;`, `\|`, `\|&`, `&`, and newlines as separators. "When the command has a `*` … Claude Code reads the rule as a wildcard pattern, not an exact match." "A redirect such as `ls > out.txt` adds a check on the target." So `Bash(<command>)` is the exact rule only for one simple command |
| A wildcard rule under auto mode | Not run. [Permission modes](https://code.claude.com/docs/en/permission-modes): "On entering auto mode, broad allow rules that grant arbitrary code execution are dropped, including … Some wildcarded paths, like `Bash($HOME/bin/*)`." The follow-up rule `Bash(<abs binary> --state-dir <abs dir> followup <project> <record> *)` has a like shape. Pass 5's live run checks whether it stands; if it is dropped, the classifier decides each follow-up |
| The state folder's reader | `internal/state/read.go:122-138` `jsonNames` reads named subfolders only (`releases/` and the rest), so a new folder such as `<state>/hook-unmatched/` is never read as state |
| Follow-up files | `docs/follow-ups/*.md`, 15 files: `# Title`, a blank line, then the body |
| The test rigs | `e2e/desk_test.go:45-60` `startDesk(t, terminal, env, script)` runs a desk under `script(1)` for a terminal or `setsid` for none, with `os.Environ()` plus `env`, so a `PATH=/usr/bin:/bin` in `env` wins. `e2e/ruling_test.go:107-132` `recordRuling` writes `desks.json` with `desk.Entry` and builds only ruling arguments. `e2e/launch_rig_test.go:351-355` `launchRig.args` refuses any verb but `check` and `session` with `t.Fatalf`. The `claude` stand-in logs `calls.log`; `--bg` makes a row; `stop` marks it stopped; controls `no-id`, `no-row`, `needs-prompt`, `fail-bg`, `stop-fails`, `lists/<n>.json`; `setRows` sets a row's state. `e2e/main_test.go:71-73` `run()` always passes `--usage-command` and `--now`, and `runWith` sets a `PATH` with no `claude` |
| The gate | `bash scripts/gate.sh` on `e604def`, 2026-11-18: gofmt, go vet, go test (e2e 40.968 s), vale, `gate: ok`, exit 0 |
| CI | `.github/workflows/gate.yml` runs the gate on each push. The watch: `gh run list --repo glw907/cc-desk --branch sessions-prompts --limit 3`, then `gh run watch <id> --repo glw907/cc-desk --exit-status` |

## Recorded searches

`grep -rnE <term>` over `cmd`, `internal`, and `e2e` (`*.go`, `*.txt`, `*.json`), 2026-11-18:

| Search | Hits | Each hit |
| --- | --- | --- |
| `newPromptCmd\|prompt hook` | 6 | `prompt.go:9`: the stub, replaced (task 2) and given `yes` and `no` (task 5). `root.go:120`: gains `newRulesCmd` and `newFollowupCmd` (tasks 4, 5). `launch_test.go:68`, `settings.go:72`, `e2e/launch_test.go:148-149`: the hook command line, left as is |
| `RelaunchNote\|\.Relaunch\b` | 12 | `launcher.go:242`, `prompt.go:26,29,30,42`, `launch_test.go:103`, `relaunch.txt:1`: the relaunch note, left; the answer note sits beside it (task 3). `carryout.go:49,88,89`, `decide_test.go:69,153`: the `sessions.Relaunch` decision constant, dismissed |
| `causeText` | 3 | `launcher.go:149,150,242`: left; the answer relaunch's cause is written to the record, and its prompt comes from the answer note (task 3) |
| `launch.Verbs(\|func Verbs` | 2 | `launcher.go:221`: gains the follow-up command (task 3). `settings.go:42`: left |
| `.guard(` | 1 | `ruling.go:72`: left calling the guard with the ruling prefix (task 1) |
| `StopPermissionPrompt\|LabelPermissionPrompt` | 12 | `decide.go:299-300`: gains the detail (task 2). `classify.go:17,50`, `snapshot/sessions.go:56`, `types.go:32,42`: left. Tests `snapshot/sessions_test.go:29`, `decide_test.go:120`, `state/sessions_test.go:70`, `checkrun_test.go:109,118`: left |
| `AllowRules\|allow_rules` | 7 | `types.go:150,152`: the field the rule events change (task 1). `launcher.go:222`: left. `state/launch_test.go:31,51,57`, `e2e/launch_test.go:29`: left |
| `followup\|follow-up\|FollowUp` | 0 | none; every name is new |
| `\.Verbs\|Relaunch\.` | 5 | `design.txt:12`, `pass.txt:12`, `conductor.txt:12`: each gains a follow-up line below the verbs (task 3). `launcher.go:221`, `relaunch.txt:1`: as above |
| `Prompts\|"prompts"` | 5 | `types.go:89`: the field the hook appends to. `check.go:262`: left. `launch.go:102`: left. `snapshot/sessions.go:32,92`: left |
| `desk.Caller()` | 1 | `ruling.go:33` (cmd): left; the new verbs call it the same way, before the lock |
| `LiveSessions(` | 7 | `check.go:416`: the definition, which the hook and the answer use. `launcher.go:141,195`, `check.go:444`, `decide.go:96,110`, `checkrun.go:45`: left |
| `WriteSettings(` | 4 | `settings.go:61`: gains the follow-up rule. `launcher.go:222`, `launch_test.go:43` and `:75`: change with the signature. `launch_test.go:58`, which counts the allow list as `1+len(SessionReasons)`, changes to count the follow-up rule too: the one earlier test this pass edits |
| `^func Test(RecordPrompt\|AnswerPrompt\|AddRule\|RemoveRule\|StateGuard\|PromptHook\|PromptYes\|PromptNo\|PromptGuard\|Followup)` | 0 | none; the new tests' names are free. `TestRules*` (`internal/state/rules_test.go:55-207`) and `TestGuard` (`e2e/guard_test.go:13`) exist, so no new test takes those names |

## File list

| File | Task | Change |
| --- | --- | --- |
| `internal/state/ruling.go` | 1 | The guard takes the refusal prefix; the three ruling constants keep their bytes |
| `internal/state/prompt.go` (new), `prompt_test.go` (new) | 1 | `RecordPrompt`, `AnswerPrompt`, `AddRule`, `RemoveRule`, and the exported guard check |
| `e2e/testdata/prompts/hook-bash.json` (new) | 2 | The documented hook input, verbatim |
| `cmd/cc-desk/prompt.go` | 2, 5 | The hook records; `prompt yes` and `prompt no` |
| `internal/sessions/decide.go`, `decide_test.go` | 2 | A prompt stop's detail names the last unanswered prompt |
| `internal/launch/settings.go`, `prompt.go`, `templates/*.txt`, `templates/answer.txt` (new), `launch_test.go` | 3 | The follow-up rule and line; the answer note and its template |
| `cmd/cc-desk/launcher.go` | 3 | Passes the follow-up command; `launchRequest` carries the answer note |
| `cmd/cc-desk/followup.go` (new), `root.go` | 4 | `followup`, and the shared writer the answer uses |
| `cmd/cc-desk/rules.go` (new), `root.go` | 5 | `rules add` and `rules remove` |
| `e2e/prompt_test.go` (new) | 6 | Done item 4, with its own argument helper |
| `docs/STATUS.md`, `docs/follow-ups/*.md` | 7 | The close |

## What the pass builds

### The hook

`cc-desk prompt hook` reads its input to the end and decodes it leniently (`encoding/json`,
unknown fields ignored, never `decodeStrict`). It takes `session_id`, `tool_name`, `tool_input`,
and `permission_suggestions`. An input that doesn't parse, or has an empty `session_id` or
`tool_name`, is a bad input: it records nothing. The prompt's command is `tool_input.command`,
else `tool_input.file_path`, else `tool_input.url`, else empty. The suggested rule is the first
`addRules` entry's first rule, written `toolName(ruleContent)`, or `toolName` with no content.
It is recorded and never applied.

It calls `state.Commit` directly with `state.LockWait`, never the cmd's `f.commit`, so it takes
no usage reading. Its event, `RecordPrompt`, finds the open record whose `session_id` matches,
appends `Prompt{Time, Tool, Command, SuggestedRule}`, and sets `noSettle`, so the commit starts
no waiting pass and appends no forecast line. `RecordPrompt` refuses an empty tool or a zero
time itself. The hook always exits 0 and prints nothing to stdout. A failure of any kind (bad
input, a lock timeout, a failed commit) goes to stderr as one line. When the input parses but no
open record matches, it writes the raw input to `<state>/hook-unmatched/<UTC time>-<pid>.json`
for pass 5 to read, and records nothing.

The check's prompt stop gains a detail: `<tool>: <command>` of the record's last unanswered
prompt, or `<tool>` with no command. A prompt stop with no prompt recorded keeps an empty
detail. If this runs past its minutes, it is filed as a follow-up instead; item 4 doesn't need
it.

### The answer

`cc-desk prompt yes <project> <record> --follow-ups <dir> [--rule <rule>]` and `cc-desk prompt no
<project> <record> --follow-ups <dir>`. The record is `conductor`, `round:<name>`, or
`pass:<name>`, as `RecordLaunch` takes it. Each answers the record's last prompt with no answer.
They take the same hidden stand-in flags as `session start`.

In order:

1. Read the caller with `desk.Caller()`, before any lock.
2. Take `check.lock`, waiting up to 60 s, and hold it for the whole run, so no check sees the
   gap between the stop and the launch and makes a counted relaunch.
3. Load the state and refuse before any side effect when any of these holds. The first match
   wins.
   1. The guard fails. The refusal names the test, with the prefix `cannot answer the prompt: `.
   2. The record has no open session record (one with no `ended` time).
   3. The record has no unanswered prompt.
   4. The record isn't stopped on `permission_prompt`: for a pass, the pass's stop; for a
      conductor or design writer, the session record's `StopReason`. A building pass, or a
      record stopped for anything else, is refused, so an old prompt carried over a relaunch
      never stops a working session.
   5. For a yes: the rule is empty or malformed. With no `--rule`, the rule is
      `Bash(<command>)` only when the tool is `Bash` and the command is one simple command: one
      line, with none of `&&`, `||`, `;`, `|`, `&`, `*`, `>`, `<`, `$(`, or a backtick. Any
      other tool or command needs `--rule`, and the refusal shows the suggested rule.
   6. For a yes: the rule starts with `Bash(gh pr merge` or `Bash(approval-queue land`. A deny
      rule wins over an allow rule, so that yes could never take; the refusal says to answer no.
   7. `--follow-ups` is missing or not an existing folder.
4. Read the list. Run `claude stop <id>` only on a live row, then read the list again. A row
   still live refuses with "not relaunched": nothing is committed, and exit 2.
5. Commit one event, `AnswerPrompt`, which runs steps 3.1 to 3.6 again inside the lock and then:
   - sets the prompt's answer to `yes` or `no`;
   - on a yes, adds the rule to the release's `allow_rules` when it isn't there;
   - for a pass, resumes it through `StartPass`. A conductor or design writer needs no resume:
     `RecordLaunch` drops the old record's stop.
6. Launch through `Launch` with `launchRequest{Kind: LaunchRelaunch, Cause: "permission_yes"` or
   `"permission_no", Counted: false, Replaces: <short id>, Answer: &launch.AnswerNote{...}}`.
   `Launch` sets `p.Answer` in place of `p.Relaunch` when `Answer` is set. The launch loads the
   state fresh, so its settings file holds the new rule. A failed launch goes through
   `failedRelaunch`, as the check's does, and the verb exits 2 once step 7 has run.
7. Write the follow-up file for the prompt into `--follow-ups`. A failure here prints the error
   and exits 1, and undoes nothing.

The answer relaunch's prompt adds a paragraph from `templates/answer.txt` in place of
`relaunch.txt`. On a yes it says the session was ended to add a rule, names the rule, says the
relaunch doesn't count toward the two, and tells it to carry on from the branch. On a no it names
the refused command, tells the session not to run it again, and tells it to find another way or
report a stop. The words are the builder's.

The prompt's follow-up file is `<dir>/prompt-<project>-<record with : as ->-<prompt time as
20061102T150405Z>.md`. Its title names the tool and command, and its body gives the project, the
record, the session name, the time, the suggested rule, the answer, the rule added, and any
older unanswered prompts. When the file exists with the same bytes, the write is a success; with
other bytes, it fails.

### The rules

`cc-desk rules add <project> <rule>` and `cc-desk rules remove <project> <rule>` run the guard
(prefix `cannot change the allow rules: `) and change the live release's `allow_rules` in one
commit through `f.commit`. A rule is a tool name of letters, digits, and `_`, optionally followed
by `(`, a non-empty content, and a closing `)` at the end. An add of a rule already there, or a
remove of one that isn't, is refused. A change takes effect at the session's next launch, since
each launch writes its settings fresh. `AddRule` is the event the yes uses too.

### The follow-up verb

`cc-desk followup <project> <record> <title> --body-file <path>` writes
`<record's worktree>/docs/follow-ups/<slug>.md` holding `# <title>`, a blank line, the body, and
a final newline. The slug is the title in lower case, with each run of characters other than
`a-z` and `0-9` made one `-`, trimmed of `-`, and cut to 60 bytes. These are refused: an empty
slug, an empty body, a record with no open session record, a body file that can't be read, and
a body file whose resolved path (after `filepath.EvalSymlinks`) lies outside both the record's
worktree and the folder in the caller's `CLAUDE_JOB_DIR`, when that is set. The last keeps the
wildcard rule from copying a file such as `~/.local/secrets` into a file that rides a PR. The
verb makes `docs/follow-ups/` when missing and opens the file with `O_EXCL`, so a name that
exists is refused with the path. It runs no guard: it writes only inside the record's own
worktree.

Each launched session's settings gain one rule,
`Bash(<binary> --state-dir <state> followup <project> <record> *)`, and each template gains one
line below the report verbs naming the command and saying to write the body to a file in the job
`tmp` folder first. The line is a new `Prompt.FollowUp` field, never an entry in `.Verbs`, and
never holds the word "relaunch".

## Acceptance

Copied from the design's "What done looks like", items 4 and 8, word for word.

4. **Prompts.** The hook command, given a `PermissionRequest` input recorded from a real session, adds the prompt to the release file with its tool and command, and prints nothing. A yes from a desk session adds the rule to the release file. The log then shows the old session ended, a settings file written after the yes that holds the new rule, and a launch naming that file. The relaunch count is unchanged. A no does the same with no new rule, and the relaunch prompt names the refused command. The same yes from a background session is refused. A follow-up file for the prompt exists.

8. **The gate.** `bash scripts/gate.sh` passes, with no skipped test. Items 6 and 7 are not Go tests. Their script's output is in the round report, and a pass that leaves it out isn't done.

Item 4's "recorded from a real session" is amended for this pass: the fixture is the hooks
reference's documented example, verbatim, as the planner prompt directs when the auto-mode entry
doesn't cover the probe. The PR says so under `## Acceptance`, and pass 5's live run records a
real input (task 7's follow-up). Items 6 and 7 are pass 5's.

| Item | Task | Check | On `e604def` |
| --- | --- | --- | --- |
| 4, the hook | 2, 6 | `go test ./e2e/ -run TestPromptHook -count=1 -v` prints `--- PASS: TestPromptHook` | The stub: exit 0, no output, nothing written to the state folder. The test doesn't exist |
| 4, the yes | 5, 6 | `go test ./e2e/ -run TestPromptYes -count=1 -v` prints `--- PASS: TestPromptYes` | `unknown command "yes" for "cc-desk prompt"`, exit 1 |
| 4, the no | 5, 6 | `go test ./e2e/ -run TestPromptNo -count=1 -v` prints `--- PASS: TestPromptNo` | `unknown command "no" for "cc-desk prompt"`, exit 1 |
| 4, the refusal | 1, 6 | `go test ./e2e/ -run TestPromptGuard -count=1 -v` prints `--- PASS: TestPromptGuard` | The verb doesn't exist |
| 4, the follow-up file | 4, 6 | `TestPromptYes` and `TestPromptNo` find the file; `go test ./e2e/ -run TestFollowup -count=1 -v` prints `--- PASS: TestFollowup` | `unknown command "followup" for "cc-desk"`, exit 1 |
| 8 | 7 | `bash scripts/gate.sh`, bare, ends `gate: ok`; `go test ./... -count=1 -v 2>&1 \| grep -c -- '--- SKIP'` prints 0 | `gate: ok`, exit 0 |

`e2e/prompt_test.go` holds its own helper that builds the five stand-in flags as
`launchRig.args` does, for `prompt` and `rules` verbs, without editing `args`. It writes
`desks.json` as `recordRuling` does, and passes `PATH=/usr/bin:/bin` in `startDesk`'s `env`, so
a missed stand-in flag can't reach the real `claude`. What each test asserts:

- `TestPromptHook`: a pass launched through the rig. The fixture's `session_id` is set to the
  record's `session_id`. The hook runs through `exec` with exactly the settings file's hook
  command line and `PATH=/usr/bin:/bin`, reading the fixture on stdin. Stdout is empty, exit 0,
  `usage.json` is unchanged, and the release file's record holds one prompt with tool `Bash`,
  the command, and the suggested rule. A second input with an unknown `session_id` records
  nothing and leaves one file under `hook-unmatched/`. An input that doesn't parse, and one
  whose `session_id` matches but has no `tool_name`, each exit 0 with stdout empty and record
  nothing, and a `snapshot` run after them exits 0.
- `TestPromptYes`: the record is seeded with one counted relaunch first. After the hook,
  `setRows` gives the row state `blocked` and `waitingFor` `"permission prompt"`, a check stops
  the pass on the prompt, and the snapshot shows it. The test reads the settings file and
  asserts the rule is absent, and notes the wall time. `prompt yes` under the desk exits 0. The
  release's `allow_rules` holds `Bash(<command>)` and the prompt's answer is `yes`. `calls.log`
  shows, in order, a list read, `stop <old id>`, a list read, and one `--bg` whose `--settings`
  path names `<state>/launch/<name>.json`. That file's mtime is after the noted wall time and
  its allow list holds the rule. The record's `Counted()` is 1 before and after, its last
  relaunch has cause `permission_yes` and `counted` false, and the pass is building. The
  follow-up file exists under `--follow-ups`.
- `TestPromptNo`: the same with `prompt no`. `allow_rules` is unchanged, the answer is `no`, and
  the relaunch prompt holds the refused command.
- `TestPromptGuard`: the same `prompt yes` run with no terminal, and run with an unregistered
  `CLAUDE_CODE_SESSION_ID`, is refused with its test named. The state folder is byte for byte
  the same after, and `calls.log` holds no `stop` and no new `--bg`. `rules add` gets the same
  two refusals.
- `TestFollowup`: the verb writes the file in the record's worktree with the title and body. A
  second run with the same title is refused, naming the path, and a body file outside the
  worktree and the job folder is refused.

## Tasks

Each task commits on its own, runs `bash scripts/gate.sh` bare before the commit, and is pushed.
Each accepted task's report names the gate check's verdict on its pushed commit. No task holds
code. Each test is table-driven where it has cases, and every Go file follows
`standards:go-conventions`.

### Task 1: the state events (20 minutes)

- Outcome: the prompt, answer, and rule events exist and are tested in `internal/state`.
- Files: `internal/state/ruling.go`, `internal/state/prompt.go` (new), `prompt_test.go` (new).
- Class: `runner`.
- Check: `go test ./internal/state/ -run 'TestRecordPrompt|TestAnswerPrompt|TestAddRule|TestRemoveRule|TestStateGuard' -count=1 -v`
  prints `--- PASS:` for each of the five names, and `go test ./internal/state/ ./e2e/ -run Ruling -count=1`
  passes.
- Test, table-driven:
  - `RecordPrompt(sessionID string, p Prompt)` appends to the one open record with that
    `session_id`, reports whether it found one, refuses an empty tool or a zero time, and sets
    `noSettle`.
  - `AnswerPrompt(c Caller, project, record, answer, rule string)` follows "The answer" steps 3.1
    to 3.6 and step 5, with a case for each refusal (including a building pass, a record
    stopped on `sign_in_expired`, a compound command, a redirect, and a deny-rule yes).
  - `AddRule(c, project, rule)` and `RemoveRule(c, project, rule)` follow "The rules".
  - An exported `(*State).Guard(c Caller, prefix string) error` gives the cmd its pre-check.
  - Each guard refusal comes back with its event's prefix, and the three ruling constants and
    the "already answered" refusal keep their bytes: `ruling_test.go` passes unchanged.

### Task 2: the hook verb, its fixture, and the stop detail (15 minutes, a pair with task 3)

- Outcome: `prompt hook` records as "The hook" says, the fixture is in place, and a prompt stop
  names its prompt.
- Files: `e2e/testdata/prompts/hook-bash.json` (new), `cmd/cc-desk/prompt.go`,
  `internal/sessions/decide.go`, `decide_test.go`.
- Class: `runner`.
- Check: `python3 -I -c 'import json,sys; d=json.load(open(sys.argv[1])); print(d["hook_event_name"], d["tool_name"], d["tool_input"]["command"])' e2e/testdata/prompts/hook-bash.json`
  prints `PermissionRequest Bash rm -rf node_modules`, and `go test ./internal/sessions/ -count=1`
  passes with a new case whose record holds one unanswered prompt and whose stop detail reads
  `Bash: <command>`.
- Test: the e2e test is task 6's `TestPromptHook`. The unit case sits beside the old
  `decide_test.go:120` case, which stays.

### Task 3: the launch package (15 minutes, a pair with task 2)

- Outcome: the settings hold the follow-up rule, the templates name the follow-up command, and
  an answer note fills `answer.txt` through `launchRequest`.
- Files: `internal/launch/settings.go`, `prompt.go`, `templates/pass.txt`, `design.txt`,
  `conductor.txt`, `templates/answer.txt` (new), `launch_test.go`, `cmd/cc-desk/launcher.go`.
- Class: `runner`.
- Check: `go test ./internal/launch/ -count=1 -v` passes, with new cases for the follow-up rule,
  the follow-up line in each kind's prompt, and both answer notes. `launch_test.go:99` and
  `:108`, and `go test ./e2e/ -run 'TestLaunch|TestRecovery' -count=1`, pass unchanged.
- Test: `WriteSettings` gains the follow-up rule as one more argument; every caller passes it
  (`launcher.go:222`, `launch_test.go:43`, `:75`). `launch_test.go:58` counts the new rule.
  `Prompt.Answer *AnswerNote{Yes bool; Command, Rule string}` and `Prompt.FollowUp string`.
  `launchRequest.Answer *launch.AnswerNote`, which `Launch` sets in place of `Relaunch`. `Fill`
  uses `answer.txt` when `Answer` is set, and refuses a prompt with both `Relaunch` and `Answer`.

The pair runs as `executor.md` "Pairs" says: two subagents, each in its own worktree branched
from task 1's commit. Task 2 touches `cmd/cc-desk/prompt.go` and `internal/sessions/`; task 3
touches `internal/launch/` and `cmd/cc-desk/launcher.go`, so no file is shared. Task 2 merges
first, then task 3, and the gate runs after each merge.

### Task 4: the follow-up verb (10 minutes)

- Outcome: `followup` writes as "The follow-up verb" says, and exports the writer task 5's step
  7 shares.
- Files: `cmd/cc-desk/followup.go` (new), `cmd/cc-desk/root.go`.
- Class: `ordinary`.
- Check: `go test ./e2e/ -run TestFollowup -count=1 -v` prints `--- PASS: TestFollowup`.
- Test: `TestFollowup` as `## Acceptance` lists it, plus unit cases for the slug.

### Task 5: the answer and rules verbs (25 minutes)

- Outcome: `prompt yes`, `prompt no`, `rules add`, and `rules remove` run as "The answer" and
  "The rules" say.
- Files: `cmd/cc-desk/prompt.go`, `cmd/cc-desk/rules.go` (new), `cmd/cc-desk/root.go`.
- Class: `runner`.
- Check: `cc-desk prompt yes --help`, `prompt no --help`, `rules add --help`, and
  `rules remove --help` each exit 0; the e2e checks are task 6's.
- Test: task 6. The order of "The answer" steps 1 to 7 is the contract; each step's failure row
  is standing answer 40. It runs after task 4, since both add to `root.go:120` and task 5 calls
  task 4's writer.

### Task 6: done item 4 (25 minutes)

- Outcome: the four e2e tests in `## Acceptance` pass.
- Files: `e2e/prompt_test.go` (new).
- Class: `runner`.
- Check: `go test ./e2e/ -run 'TestPrompt' -count=1 -v` prints `--- PASS` for `TestPromptHook`,
  `TestPromptYes`, `TestPromptNo`, and `TestPromptGuard`.
- Test: as `## Acceptance` sets out, with the file's own argument helper. Every call to a verb
  that launches passes the stand-in flags (pass 2's answer 36). No old fixture, rig function, or
  control changes (round plan answer 16); a new one goes beside the old ones.

### Task 7: close (15 minutes)

- Outcome: the PR is ready, with its review accepted and CI green.
- Files: `docs/STATUS.md` (below the round-report block), and one file each under
  `docs/follow-ups/`: pass 5's live record of a hook input and its `session_id` match; the
  follow-up wildcard rule under auto mode; and the desk-registration gap (Decisions).
- Class: `ordinary`.
- Check: the PR body's `## Acceptance` evidence; the `gate` check on the last commit.
- Test: none of its own. One whole-branch `standards:diff-reviewer` read (class `runner`),
  folded. A Low finding from a later delta read is folded once, checked by the executor's own
  diff read and CI, or filed; it never opens another review round.

Path: task 1, tasks 2 and 3 as a pair, task 4, task 5, task 6, task 7: 20 + 15 + 10 + 25 + 25 +
15 = 110 minutes. The tasks sum to 125. At the last rounds' pace of about 0.6, the path is about
66 minutes. A task that runs long is the first sign to watch.

The CI watch, run bare after each push: `gh run list --repo glw907/cc-desk --branch
sessions-prompts --limit 3`, then `gh run watch <id> --repo glw907/cc-desk --exit-status` on the
newest.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes (4 replaced, 1 and 9 changed by
"Until cc-desk runs", 6 changed) and its answers 13 to 20. Pass 1's 23 to 29 and pass 2's 30 to
38 hold. The last round's 19 to 22 hold too: no Go test skips; a field or verb name is the
builder's; and a run whose path passes its ceiling finishes the task in flight, pushes, and
writes the stop. For this pass the ceiling is 120 minutes, in place of answer 22's 180. This
pass adds:

- **39. The path passes 120 minutes.** Finish the task in flight, push, and write the stop under
  `## Stop` in STATUS. The stop names the tasks done, the acceptance items still unmet, and the
  tasks left, and the PR stays a draft. Before that point, when task 6 can't fit, the run still
  lands tasks 1 to 5 in order, and the close reports item 4 unmet.
- **40. A step of the answer fails.**
  - At step 2, the lock isn't free within 60 s: nothing has happened. The refusal says to run it
    again after the check, and the run exits 2.
  - At step 3, nothing has happened; the refusal names the reason and the run exits 2.
  - At step 4, a failed list read, a failed stop, or a row still live commits nothing. A rerun
    repeats the run.
  - At step 5, a failed commit leaves the session stopped and the prompt unanswered, and the
    record still stopped on `permission_prompt`. A rerun finds no live row, skips the stop, and
    goes on.
  - At step 6, the answer is committed and `failedRelaunch` stops the record as
    `launch_failed`; the follow-up is still written. A rerun is refused, since no prompt is
    unanswered, and the record's stop says what to do, as for any failed launch.
  - At step 7, the relaunch stands; the error names the file, and the conductor writes it by
    hand from the record.
- **41. The hook meets something it can't record.** It exits 0 with empty stdout every time.
  Never exit 2, never print a decision, and never set `onFailure`. A lock wait past
  `state.LockWait` is a failure like any other, written to stderr.
- **42. A prompt arrives for a record with prompts already unanswered.** Append it. The answer
  takes the last unanswered one; older unanswered prompts stay unanswered and show in the
  follow-up's body.
- **43. A yes for a tool other than Bash, or with a command that isn't one simple command.**
  Refused at step 3 unless `--rule` names the rule. The suggested rule is shown in the refusal
  but never applied on its own.
- **44. The prompt's follow-up folder.** Always the `--follow-ups` flag; never a default, and
  never the main checkout, where an untracked file trips `claude-worktree-busy`. Tests pass a
  `t.TempDir()`.
- **45. A test would need a desk registered by a verb.** Write `desks.json` as `recordRuling`
  does. Don't add a verb that registers a desk.
- **46. A test needs a rig change.** Add a helper in `e2e/prompt_test.go` beside the old ones;
  the one earlier test this pass edits is `internal/launch/launch_test.go:58`, for the allow
  count.

## Decisions

- [advisor] No live probe. The planner prompt allows one only if the auto-mode entry covers it,
  and it doesn't; the log also refused a spike settings file as Self-Modification on
  2026-11-18. The fixture is the documented example, and item 4's "recorded from a real
  session" is amended in `## Acceptance`.
  Call: mine (method, as the planner prompt directs).
- [advisor] The answer verb takes `check.lock`, runs the guard and every refusal before any side
  effect, stops and confirms, commits once with the checks run again, then launches uncounted.
  Each step has a failure row (answer 40).
  Call: mine (method).
- [advisor] The answer acts only on a record stopped on `permission_prompt`, so an old prompt
  carried over a relaunch never stops a working session and a record stopped for another reason
  is never relaunched by it.
  Call: mine (method).
- [advisor] The three ruling refusal constants keep their bytes, and the prompt and rules verbs
  get prefixes of their own through the guard's new prefix argument.
  Call: mine (method).
- [advisor] A yes adds `Bash(<command>)` only for one simple command, which the permissions page
  says is the exact rule. Any other command or tool needs `--rule` (answer 43). The suggested
  rule is recorded only, since it is optional and may be a wildcard.
  Call: mine (method, within "a yes adds the exact rule").
- [advisor] The hook decodes leniently, matches on `session_id` alone, commits directly with no
  usage reading and no settle, always exits 0, and keeps the raw input of a miss under
  `<state>/hook-unmatched/`. The match is unprobed.
  Call: mine (method).
- [advisor] The follow-up command has its own wildcard rule and its own prompt line, outside
  `.Verbs`, and takes the body from a file inside the worktree or the job folder. Whether auto
  mode keeps the rule is unprobed, and task 7 files it for pass 5.
  Call: mine (method).
- [advisor] The design says each prompt's follow-up is "written into the cc-desk repo's folder",
  but the binary has no way to know that folder. The answer verb takes it as the required
  `--follow-ups` flag, with no default (answer 44). The file is named from the prompt's time, so
  a rerun finds its own file.
  Call: mine (method, "Left to the builder"). Pass 5's conductor names the folder.
- [advisor] The answer relaunch gets `templates/answer.txt` through `launchRequest.Answer`, so
  `relaunch.txt` and its "n of 2" wording stay as two tests pin them. Launch-review note 5 (an
  uncounted usage relaunch reads "1 of 2") is left to its follow-up.
  Call: mine (method).
- [advisor] An answer resumes a pass through `StartPass`, which checks no limits, so an answer
  can leave three passes building. That gap is already filed as
  `docs/follow-ups/resume-skips-the-limits.md`; this pass doesn't close it.
  Call: mine (scope).
- [advisor] No verb registers a desk, so in practice no session can answer a prompt yet. Round
  plan answer 17 has pass 5's conductor register as a desk. Task 7 files this as a follow-up
  naming answer 17; this pass adds no verb.
  Call: mine (scope), flagged for the conductor in the PR body.
- The answer's cause is written as `permission_yes` or `permission_no` in the relaunch entry, and
  the entry is `counted: false`, so the count shown is unchanged.
  Call: mine (naming, standing answer 21).
- Verb names: `prompt yes|no <project> <record>`, `rules add|remove <project> <rule>`, and
  `followup <project> <record> <title> --body-file <path>`. Folder: `<state>/hook-unmatched/`.
  Call: mine (naming, standing answer 21).
- Builds on PR 8's flagged call: nothing. The usage relaunch stays as pass 1 decides it and pass 2
  carries it out. This pass's relaunch is a new uncounted cause on its own path, which never
  goes through `early`'s usage case.
  Call: none.
- Builds on PR 8's and PR 10's `[advisor]` lines: the launch is pass 2's `Launch`, under
  `check.lock`, with its record committed by its own commit, and `failedRelaunch` on a failure.
  Call: mine (method).

## Review findings and their folds

The plan-reviewer read of 2026-11-18, most severe first:

1. High: a prompt with an empty tool breaks every verb. Folded: `RecordPrompt` refuses it, the
   hook treats a missing `tool_name` as bad input, and `TestPromptHook` checks a `snapshot`
   after.
2. High: the cmd's commit takes a usage reading that can outrun the hook's 30 s. Folded: the
   hook commits directly with no reading and no settle; the test runs the real hook command line
   with a `PATH` holding no `claude`.
3. High: `Bash(<command>)` isn't exact for a compound command or a redirect. Folded: the default
   rule only for one simple command (step 3.5, answer 43).
4. High: the pairs shared `root.go` and hid a dependency. Folded: the follow-up verb runs before
   the answer verbs, and the remaining pair touches no shared file, in their own worktrees.
5. Medium: the probe went against the planner prompt and was broken as written. Folded: no
   probe.
6. Medium: the rigs can't run `prompt yes`. Folded: an argument helper in `e2e/prompt_test.go`,
   `PATH` in the desk's `env`, and `setRows` for the blocked row.
7. Medium: "a settings file written after the yes" wasn't observable. Folded: before and after
   reads of the file, a wall-time mtime check, and the call order in `calls.log`.
8. Medium: the answer note had no way into `Launch`. Folded: `launchRequest.Answer`.
9. Medium: "live session" was undefined. Folded: "open record" means no `ended` time.
10. Medium: an old prompt could stop a working session. Folded: step 3.4 accepts only a record
    stopped on `permission_prompt`.
11. Medium: auto mode may drop the follow-up rule, and `--body-file` read any path. Folded: an
    unprobed mechanism row, a follow-up, and a body-path refusal with a test.
12. Medium: a check holds the lock past 10 s, and two failure rows were missing. Folded: a 60 s
    wait and both rows in answer 40.
13. Medium: task 1's check matched existing tests. Folded: new names and a check for each.
14. Low: `launch_test.go:58` breaks and the cited line was wrong. Folded: the search row, task
    3, and answer 46.
15. Low: a resume skips the limits. Folded: the Decisions line citing its follow-up.
16. Low: "Reviewed:" pointed at a missing list. Folded: this list.
17. Low: the overrun answer cited answer 14. Folded: answer 39, with the unmet-bar case.
18. Low: task 8 lacked two process lines. Folded: the CI verdict per task and the delta-read
    floor.
19. Low: the `Counted()` check proved nothing from 0. Folded: a counted relaunch seeded first.
