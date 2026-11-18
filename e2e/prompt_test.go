//go:build unix

package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/state"
)

const (
	hookCommand  = "rm -rf node_modules"
	hookRule     = "Bash(rm -rf node_modules)"
	promptDesk   = "e2e-prompt-desk"
	hookFixture  = "testdata/prompts/hook-bash.json"
	hookFixtureS = `"session_id":"abc123"`
)

// hookInput gives the documented fixture with its session_id replaced.
func hookInput(t *testing.T, sessionID string) string {
	t.Helper()
	b, err := os.ReadFile(hookFixture)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(hookFixtureS)) {
		t.Fatalf("%s no longer holds %s", hookFixture, hookFixtureS)
	}
	return strings.Replace(string(b), hookFixtureS, `"session_id":`+strconv.Quote(sessionID), 1)
}

// runHook runs the PermissionRequest hook command line the pass's settings
// file names, as claude would, with input on its standard input.
func (r launchRig) runHook(t *testing.T, input string) result {
	t.Helper()
	var s struct {
		Hooks struct {
			PermissionRequest []struct{ Hooks []struct{ Command string } }
		}
	}
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "launch/proj-a.json")), &s); err != nil {
		t.Fatal(err)
	}
	if len(s.Hooks.PermissionRequest) != 1 || len(s.Hooks.PermissionRequest[0].Hooks) != 1 {
		t.Fatalf("settings hooks = %+v, want one PermissionRequest command", s.Hooks)
	}
	cmd := exec.Command("sh", "-c", s.Hooks.PermissionRequest[0].Hooks[0].Command)
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	cmd.Stdin = strings.NewReader(input)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	code := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	return result{out.String(), errOut.String(), code}
}

// asDesk runs a verb that can launch with every stand-in flag set, as a
// desk's command with a terminal, or without one when terminal is false. The
// desk's own session id is registered; sessionID is the one the verb sees.
// beforeRun is called once desks.json is in place and just before the verb
// runs.
func (r launchRig) asDesk(t *testing.T, terminal bool, sessionID, now string, beforeRun func(), verb ...string) result {
	t.Helper()
	args := append(globalArgs(t, r.state, now), verb...)
	args = append(args,
		"--claude-command", r.claude,
		"--systemd-run-command", r.systemd,
		"--jobs-dir", r.jobs,
		"--second-read-wait", "1ms",
		"--launch-row-wait", "50ms",
	)
	out := t.TempDir()
	script := `( setsid -w "$BIN" "$@" > "$OUT/stdout" 2> "$OUT/stderr"; echo $? > "$OUT/code" ) &`
	script = "set -- " + shellWords(args) + "\n" + script
	env := []string{"BIN=" + binary, "OUT=" + out, "CLAUDE_CODE_SESSION_ID=" + sessionID, "PATH=/usr/bin:/bin"}
	pid, release, wait := startDesk(t, terminal, env, script)

	entry, err := desk.Entry(promptDesk, pid)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state.Desks{Format: state.Format, Desks: []state.Desk{entry}})
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, r.state, "desks.json", string(data)+"\n")
	if beforeRun != nil {
		beforeRun()
	}
	release()
	wait()
	code, err := strconv.Atoi(strings.TrimSpace(readFile(t, filepath.Join(out, "code"))))
	if err != nil {
		t.Fatalf("the verb left no exit code: %v", err)
	}
	return result{readFile(t, filepath.Join(out, "stdout")), readFile(t, filepath.Join(out, "stderr")), code}
}

// counted gives how many of a session record's relaunches count.
func counted(sess map[string]any) int {
	n := 0
	rel, _ := sess["relaunches"].([]any)
	for _, x := range rel {
		if x.(map[string]any)["counted"] == true {
			n++
		}
	}
	return n
}

// promptsOf gives the prompts on pass a's session record.
func (r launchRig) promptsOf(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	ps, _ := r.sessionOf(t, "a")["prompts"].([]any)
	for _, p := range ps {
		out = append(out, p.(map[string]any))
	}
	return out
}

// settingsAllow reads the allow list of pass a's settings file.
func (r launchRig) settingsAllow(t *testing.T) []string {
	t.Helper()
	var s struct{ Permissions struct{ Allow []string } }
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "launch/proj-a.json")), &s); err != nil {
		t.Fatal(err)
	}
	return s.Permissions.Allow
}

// stoppedOnPrompt launches pass a, relaunches it once for a counted cause,
// records a prompt through the hook, blocks its row on the prompt, and runs a
// check that stops the pass.
func stoppedOnPrompt(t *testing.T) launchRig {
	t.Helper()
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	r.check(t, minutes(t, 1))
	r.vanish(t, r.sessionOf(t, "a")["short_id"].(string))
	r.check(t, minutes(t, 2))
	r.check(t, minutes(t, 3))
	if n := counted(r.sessionOf(t, "a")); n != 1 {
		t.Fatalf("seeded %d counted relaunches, want 1", n)
	}

	got := r.runHook(t, hookInput(t, r.sessionOf(t, "a")["session_id"].(string)))
	if got.code != 0 || got.stdout != "" {
		t.Fatalf("hook: exit %d, stdout %q, stderr %q", got.code, got.stdout, got.stderr)
	}
	rows := r.readRows(t)
	rows[0]["state"], rows[0]["status"], rows[0]["waitingFor"] = "blocked", "blocked", "permission prompt"
	r.setRows(t, rows...)
	r.check(t, minutes(t, 4))
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "permission_prompt" {
		t.Fatalf("pass = %v, want stopped on permission_prompt", p)
	}
	return r
}

func TestPromptHook(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	sessionID := r.sessionOf(t, "a")["session_id"].(string)
	usageBefore := readFile(t, filepath.Join(r.state, "usage.json"))

	got := r.runHook(t, hookInput(t, sessionID))
	if got.code != 0 || got.stdout != "" {
		t.Fatalf("hook: exit %d, stdout %q, stderr %q; want exit 0 and nothing on stdout", got.code, got.stdout, got.stderr)
	}
	if now := readFile(t, filepath.Join(r.state, "usage.json")); now != usageBefore {
		t.Errorf("usage.json changed:\n%s\nwas:\n%s", now, usageBefore)
	}
	ps := r.promptsOf(t)
	if len(ps) != 1 {
		t.Fatalf("prompts = %v, want one", ps)
	}
	if ps[0]["tool"] != "Bash" || ps[0]["command"] != hookCommand || ps[0]["suggested_rule"] != hookRule {
		t.Errorf("prompt = %v, want Bash, %q, and rule %q", ps[0], hookCommand, hookRule)
	}

	unmatched := func() int {
		entries, _ := os.ReadDir(filepath.Join(r.state, "hook-unmatched"))
		return len(entries)
	}
	got = r.runHook(t, hookInput(t, "no-such-session"))
	if got.code != 0 || got.stdout != "" {
		t.Errorf("unknown session: exit %d, stdout %q; want exit 0 and nothing on stdout", got.code, got.stdout)
	}
	if n := unmatched(); n != 1 {
		t.Errorf("%d files under hook-unmatched, want 1", n)
	}
	if n := len(r.promptsOf(t)); n != 1 {
		t.Errorf("%d prompts after the unknown session, want 1", n)
	}

	for name, in := range map[string]string{
		"not JSON": "this is not json",
		"no tool":  `{"session_id":` + strconv.Quote(sessionID) + `,"hook_event_name":"PermissionRequest"}`,
	} {
		got = r.runHook(t, in)
		if got.code != 0 || got.stdout != "" {
			t.Errorf("%s: exit %d, stdout %q; want exit 0 and nothing on stdout", name, got.code, got.stdout)
		}
		if n := len(r.promptsOf(t)); n != 1 {
			t.Errorf("%s: %d prompts, want 1", name, n)
		}
	}
	requireRun(t, r.do(t, workedClock, "snapshot"))
}

// followUpPath names the follow-up file the answer wrote for pass a's prompt.
func followUpPath(t *testing.T, r launchRig, dir string) string {
	t.Helper()
	ps := r.promptsOf(t)
	if len(ps) == 0 {
		t.Fatal("the record holds no prompt")
	}
	when, err := time.Parse(time.RFC3339, ps[0]["time"].(string))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "prompt-proj-pass-a-"+when.UTC().Format("20060102T150405Z")+".md")
}

// answered runs a yes or no under a desk with a terminal and checks what both
// share: the stop, the relaunch from a fresh settings file, and the file.
func answered(t *testing.T, answer string) (r launchRig, follow string) {
	t.Helper()
	r = stoppedOnPrompt(t)
	old := r.sessionOf(t, "a")["short_id"].(string)

	if needs := r.needs(t); len(needs) == 0 || !strings.Contains(mustJSON(t, needs), hookCommand) {
		t.Errorf("needs_you_now = %v, want the prompt's command", needs)
	}
	if slices.Contains(r.settingsAllow(t), hookRule) {
		t.Fatalf("settings allow %v already hold %s", r.settingsAllow(t), hookRule)
	}
	info, err := os.Stat(filepath.Join(r.state, "launch", "proj-a.json"))
	if err != nil {
		t.Fatal(err)
	}
	callsBefore := len(r.calls(t))
	launchesBefore := len(r.launches(t))
	follow = t.TempDir()

	got := r.asDesk(t, true, promptDesk, minutes(t, 5), nil, "prompt", answer, "proj", "pass:a", "--follow-ups", follow)
	if got.code != 0 {
		t.Fatalf("prompt %s: exit %d; stderr: %s", answer, got.code, got.stderr)
	}

	after := r.calls(t)[callsBefore:]
	var kinds []string
	for _, c := range after {
		kinds = append(kinds, c.Args[0])
	}
	if len(kinds) < 4 || !slices.Equal(kinds[:3], []string{"agents", "stop", "agents"}) {
		t.Fatalf("calls after the answer = %v, want agents, stop, agents, then a launch", kinds)
	}
	if after[1].Args[1] != old {
		t.Errorf("stopped %v, want %s", after[1].Args, old)
	}
	if n := len(r.launches(t)) - launchesBefore; n != 1 {
		t.Fatalf("%d launches after the answer, want 1", n)
	}
	launch := r.launches(t)[launchesBefore]
	i := slices.Index(launch.Args, "--settings")
	if i < 0 || launch.Args[i+1] != filepath.Join(r.state, "launch", "proj-a.json") {
		t.Errorf("launch args lack --settings %s: %v", filepath.Join(r.state, "launch", "proj-a.json"), launch.Args)
	}
	if now, err := os.Stat(filepath.Join(r.state, "launch", "proj-a.json")); err != nil || now.ModTime().Before(info.ModTime()) {
		t.Errorf("settings file mtime %v, %v; want it not before %v", now, err, info.ModTime())
	}

	sess := r.sessionOf(t, "a")
	if n := counted(sess); n != 1 {
		t.Errorf("%d counted relaunches, want 1 as before", n)
	}
	rel := sess["relaunches"].([]any)
	last := rel[len(rel)-1].(map[string]any)
	if last["cause"] != "permission_"+answer || last["counted"] == true {
		t.Errorf("last relaunch = %v, want permission_%s, not counted", last, answer)
	}
	if p := r.passRecord(t, "a"); p["state"] != "building" {
		t.Errorf("pass = %v, want building", p)
	}
	if ps := r.promptsOf(t); len(ps) != 1 || ps[0]["answer"] != answer {
		t.Errorf("prompts = %v, want one answered %s", ps, answer)
	}

	file := followUpPath(t, r, follow)
	body := readFile(t, file)
	for _, want := range []string{"# Bash: " + hookCommand, "Answer: " + answer, "Record: pass:a"} {
		if !strings.Contains(body, want) {
			t.Errorf("follow-up file %s lacks %q:\n%s", file, want, body)
		}
	}
	return r, follow
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPromptYes(t *testing.T) {
	r, follow := answered(t, "yes")
	rules, _ := r.release(t)["allow_rules"].([]any)
	if !slices.Contains(rules, any(hookRule)) || !slices.Contains(rules, any(launchAllow)) {
		t.Errorf("allow_rules = %v, want the release's rule and %s", rules, hookRule)
	}
	if !slices.Contains(r.settingsAllow(t), hookRule) {
		t.Errorf("settings allow %v lack %s", r.settingsAllow(t), hookRule)
	}
	if body := readFile(t, followUpPath(t, r, follow)); !strings.Contains(body, "Rule added: "+hookRule) {
		t.Errorf("follow-up file lacks the rule added:\n%s", body)
	}
}

func TestPromptNo(t *testing.T) {
	r, follow := answered(t, "no")
	rules, _ := r.release(t)["allow_rules"].([]any)
	if len(rules) != 1 || rules[0] != launchAllow {
		t.Errorf("allow_rules = %v, want only %s", rules, launchAllow)
	}
	if slices.Contains(r.settingsAllow(t), hookRule) {
		t.Errorf("settings allow %v hold %s after a no", r.settingsAllow(t), hookRule)
	}
	launches := r.launches(t)
	if p := relaunchPrompt(launches[len(launches)-1]); !strings.Contains(p, hookCommand) {
		t.Errorf("relaunch prompt lacks the refused command:\n%s", p)
	}
	if body := readFile(t, followUpPath(t, r, follow)); strings.Contains(body, "Rule added") {
		t.Errorf("follow-up file names a rule after a no:\n%s", body)
	}
}

func TestPromptGuard(t *testing.T) {
	r := stoppedOnPrompt(t)
	release := filepath.Join("releases", "proj.json")
	follow := t.TempDir()

	for name, c := range map[string]struct {
		terminal  bool
		sessionID string
	}{
		"no terminal":          {false, promptDesk},
		"unregistered session": {true, "not-a-desk-session"},
	} {
		t.Run(name, func(t *testing.T) {
			var before map[string]string
			callsBefore := len(r.calls(t))
			snap := func() { before = folder(t, r.state) }

			got := r.asDesk(t, c.terminal, c.sessionID, minutes(t, 5), snap, "prompt", "yes", "proj", "pass:a", "--follow-ups", follow)
			if got.code != 2 || !strings.Contains(got.stderr, "cannot answer the prompt: ") {
				t.Errorf("prompt yes: exit %d, stderr %q; want exit 2 with the refusal", got.code, got.stderr)
			}
			sameFolder(t, before, folder(t, r.state))
			if n := len(r.calls(t)); n != callsBefore {
				t.Errorf("%d claude calls after the refusal, want none: %v", n-callsBefore, r.kinds(t))
			}
			if entries, _ := os.ReadDir(follow); len(entries) != 0 {
				t.Errorf("%d files in the follow-up folder after a refusal", len(entries))
			}

			relBefore := readFileIn(t, r.state, release)
			got = r.asDesk(t, c.terminal, c.sessionID, minutes(t, 5), nil, "rules", "add", "proj", "Bash(ls)")
			if got.code != 2 || !strings.Contains(got.stderr, "cannot change the allow rules: ") {
				t.Errorf("rules add: exit %d, stderr %q; want exit 2 with the refusal", got.code, got.stderr)
			}
			if now := readFileIn(t, r.state, release); now != relBefore {
				t.Errorf("the release file changed after a refused rules add")
			}
		})
	}
}
