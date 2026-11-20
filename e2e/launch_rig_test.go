//go:build unix

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// launchRig is everything one launch test runs against: a git repo with a
// bare remote, a stand-in claude, a stand-in systemd-run, and the logs both
// write. No test that uses it runs the real claude or the real systemd-run.
type launchRig struct {
	dir     string // holds the stand-ins and their files
	claude  string
	systemd string
	jobs    string
	remote  string // the bare repo
	clone   string // the project's checkout
	state   string // the state folder the test passes as --state-dir
}

// The stand-in claude controls, each a file in the rig folder that a test
// creates or removes:
//
//	no-id         --bg prints nothing, so the launch falls back to the list
//	no-row        --bg adds no row, so the ids come from the job's state.json
//	needs-prompt  the new row carries needs "send a prompt to start"
//	fail-bg       --bg exits 1 after logging
//	stop-fails    stop exits 1 and changes nothing, unless stop-marks is set
//	stop-ignored  stop exits 0 and changes nothing
//	stop-marks    with stop-fails, stop first marks the row stopped, as a
//	              stop that did its work and then failed to report it
//	lists/<n>.json  the n-th list read prints this file, and a later read
//	                with no file prints rows.json
//
// Every call appends one JSON line to calls.log: its arguments and working
// folder. A list read also advances reads.log.
const launchClaudeStandIn = `#!/usr/bin/env python3
import json, os, sys

d = %q
a = sys.argv[1:]


def path(*p):
    return os.path.join(d, *p)


def has(name):
    return os.path.exists(path(name))


with open(path("calls.log"), "a") as f:
    f.write(json.dumps({"args": a, "cwd": os.getcwd()}) + "\n")


def load():
    with open(path("rows.json")) as f:
        return json.load(f)


def save(rows):
    with open(path("rows.json.tmp"), "w") as f:
        json.dump(rows, f)
    os.replace(path("rows.json.tmp"), path("rows.json"))


def bump(name):
    n = 0
    if has(name):
        with open(path(name)) as f:
            n = int(f.read() or 0)
    n += 1
    with open(path(name), "w") as f:
        f.write(str(n))
    return n


cmd = a[0] if a else ""
if cmd == "agents":
    n = bump("reads")
    p = path("lists", "%%d.json" %% n)
    if not os.path.exists(p):
        p = path("rows.json")
    with open(p) as f:
        sys.stdout.write(f.read())
elif cmd == "--bg":
    if has("fail-bg"):
        sys.stderr.write("claude: could not start the session\n")
        sys.exit(1)
    name = a[a.index("--name") + 1] if "--name" in a else ""
    n = bump("counter")
    sid = "b%%07d" %% n
    full = sid + "-0000"
    if not has("no-row"):
        row = {"pid": 4000 + n, "id": sid, "sessionId": full, "name": name,
               "cwd": os.getcwd(), "kind": "background",
               "startedAt": 1794970000000, "status": "working", "state": "working"}
        if has("needs-prompt"):
            row["needs"] = "send a prompt to start"
        rows = load()
        rows.append(row)
        save(rows)
    os.makedirs(path("jobs", sid), exist_ok=True)
    with open(path("jobs", sid, "state.json"), "w") as f:
        json.dump({"sessionId": full, "name": name}, f)
    if not has("no-id"):
        sys.stdout.write("backgrounded · %%s · %%s\n" %% (sid, name))
elif cmd == "stop":
    rows = load()
    if has("stop-ignored"):
        sys.stdout.write("stopped %%s\n" %% a[1])
        sys.exit(0)
    if has("stop-fails") and not has("stop-marks"):
        sys.stderr.write("claude: could not stop the session\n")
        sys.exit(1)
    for r in rows:
        if r["id"] == a[1]:
            r["state"], r["status"] = "stopped", "stopped"
    save(rows)
    if has("stop-fails"):
        sys.stderr.write("claude: could not stop the session\n")
        sys.exit(1)
    sys.stdout.write("stopped %%s\n" %% a[1])
elif cmd == "logs":
    pass
`

// launchSystemdStandIn logs the three flags and the working folder to
// scope.log, then runs the rest of the command line.
const launchSystemdStandIn = `#!/bin/sh
d=%q
printf 'flags=%%s %%s %%s cwd=%%s\n' "$1" "$2" "$3" "$(pwd)" >> "$d/scope.log"
if [ "$1 $2 $3" != "--user --scope -q" ]; then
  echo "systemd-run stand-in: unexpected flags: $*" >&2
  exit 64
fi
shift 3
exec "$@"
`

// newLaunchRig makes the stand-ins, a repo with a bare remote, and an empty
// state folder.
func newLaunchRig(t *testing.T) launchRig {
	t.Helper()
	dir := t.TempDir()
	r := launchRig{
		dir:     dir,
		claude:  filepath.Join(dir, "claude.py"),
		systemd: filepath.Join(dir, "systemd-run.sh"),
		jobs:    filepath.Join(dir, "jobs"),
		state:   filepath.Join(dir, "state"),
	}
	if err := os.MkdirAll(filepath.Join(dir, "lists"), 0o755); err != nil {
		t.Fatal(err)
	}
	r.write(t, "rows.json", "[]")
	if err := os.WriteFile(r.claude, fmt.Appendf(nil, launchClaudeStandIn, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(r.systemd, fmt.Appendf(nil, launchSystemdStandIn, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	r.remote = filepath.Join(dir, "remote.git")
	r.clone = filepath.Join(dir, "project")
	git(t, dir, "init", "--bare", "-b", "main", r.remote)
	git(t, dir, "clone", r.remote, r.clone)
	git(t, r.clone, "checkout", "-b", "main")
	commitFile(t, r.clone, "README.md", "hello\n", "Start the project")
	git(t, r.clone, "push", "-u", "origin", "main")
	return r
}

// git runs a real git command in dir and returns its trimmed output.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// commitFile writes a file in the repo and commits it.
func commitFile(t *testing.T, repo, name, content, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", name)
	git(t, repo, "commit", "-m", msg)
}

// pushAnother clones the remote afresh, commits a file there, and pushes it,
// so the project's own checkout is behind the remote.
func (r launchRig) pushAnother(t *testing.T, name string) string {
	t.Helper()
	other := filepath.Join(t.TempDir(), "other")
	git(t, r.dir, "clone", r.remote, other)
	commitFile(t, other, name, "more\n", "Add "+name)
	git(t, other, "push", "origin", "main")
	return git(t, other, "rev-parse", "HEAD")
}

func (r launchRig) write(t *testing.T, rel, content string) {
	t.Helper()
	p := filepath.Join(r.dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// control turns a stand-in control on, or off with on false.
func (r launchRig) control(t *testing.T, name string, on bool) {
	t.Helper()
	p := filepath.Join(r.dir, name)
	if on {
		r.write(t, name, "")
		return
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

// setRows replaces the list every read prints, as the rows' JSON objects.
func (r launchRig) setRows(t *testing.T, rows ...map[string]any) {
	t.Helper()
	if rows == nil {
		rows = []map[string]any{}
	}
	b, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	r.write(t, "rows.json", string(b))
}

// readRows returns the rows the stand-in holds.
func (r launchRig) readRows(t *testing.T) []map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, "rows.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

// vanish removes the row with the short id, so the next list read misses it.
func (r launchRig) vanish(t *testing.T, id string) {
	t.Helper()
	var keep []map[string]any
	for _, row := range r.readRows(t) {
		if row["id"] != id {
			keep = append(keep, row)
		}
	}
	r.setRows(t, keep...)
}

// setList scripts the n-th list read, counting from 1.
func (r launchRig) setList(t *testing.T, n int, rows ...map[string]any) {
	t.Helper()
	if rows == nil {
		rows = []map[string]any{}
	}
	b, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	r.write(t, filepath.Join("lists", fmt.Sprintf("%d.json", n)), string(b))
}

// call is one line of calls.log.
type call struct {
	Args []string `json:"args"`
	Cwd  string   `json:"cwd"`
}

// calls returns the stand-in claude's calls in order.
func (r launchRig) calls(t *testing.T) []call {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, "calls.log"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []call
	for line := range strings.SplitSeq(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		var c call
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			t.Fatalf("calls.log line %q: %v", line, err)
		}
		out = append(out, c)
	}
	return out
}

// launches returns the calls that started a session.
func (r launchRig) launches(t *testing.T) []call {
	t.Helper()
	var out []call
	for _, c := range r.calls(t) {
		if len(c.Args) > 0 && c.Args[0] == "--bg" {
			out = append(out, c)
		}
	}
	return out
}

// kinds returns the first argument of each call: agents, --bg, stop, or logs.
func (r launchRig) kinds(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, c := range r.calls(t) {
		if len(c.Args) > 0 {
			out = append(out, c.Args[0])
		}
	}
	return out
}

// scopeLog returns scope.log, empty when the stand-in never ran.
func (r launchRig) scopeLog(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, "scope.log"))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// args builds the command line for a check or session verb with every
// stand-in flag set. A verb that could launch a real session without them is
// refused here, so no test can reach the real claude or systemd-run.
func (r launchRig) args(t *testing.T, verb ...string) []string {
	t.Helper()
	if len(verb) == 0 || (verb[0] != "check" && verb[0] != "session") {
		t.Fatalf("launchRig.args builds only check and session commands, got %v", verb)
	}
	return append(verb,
		"--claude-command", r.claude,
		"--systemd-run-command", r.systemd,
		"--jobs-dir", r.jobs,
		"--second-read-wait", "1ms",
		"--launch-row-wait", "50ms",
	)
}

// run runs the binary on the rig's state folder with the stand-ins.
func (r launchRig) run(t *testing.T, now string, verb ...string) result {
	t.Helper()
	return run(t, r.state, now, r.args(t, verb...)...)
}

// sleepingGhStandIn wraps the gh stand-in. When the file gh-sleeps exists, a
// `pr merge` call starts a child sleep that holds stdout, records its pid in
// gh-child.pid, and waits. Every other call runs the real stand-in, with git
// found on the plain system path so the sleeping git below does not see it.
const sleepingGhStandIn = `#!/bin/sh
d=%q
if [ -e "$d/gh-sleeps" ] && [ "$1 $2" = "pr merge" ]; then
  sleep 120 &
  echo $! > "$d/gh-child.pid"
  wait
  exit 1
fi
PATH=/usr/bin:/bin exec %q "$@"
`

// sleepingGitStandIn is a git for the front of the binary's PATH. When the
// file git-sleeps exists and its text is one of the call's arguments, the
// call starts a child sleep holding stdout, records the pid in git-child.pid,
// and waits. Every other call goes to the real git by absolute path.
const sleepingGitStandIn = `#!/bin/sh
d=%q
if [ -e "$d/git-sleeps" ]; then
  verb=$(cat "$d/git-sleeps")
  for a in "$@"; do
    if [ "$a" = "$verb" ]; then
      sleep 120 &
      echo $! > "$d/git-child.pid"
      wait
      exit 1
    fi
  done
fi
exec %q "$@"
`

// sleepers writes the sleeping gh wrapper around the gh stand-in at realGh,
// and the sleeping git in a folder of its own. It returns the gh path and the
// folder to put first on the binary's PATH. The controls are the files
// gh-sleeps and git-sleeps, which a test creates with r.write.
func (r launchRig) sleepers(t *testing.T, realGh string) (gh, pathDir string) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gh = filepath.Join(r.dir, "gh-sleeper.sh")
	if err := os.WriteFile(gh, fmt.Appendf(nil, sleepingGhStandIn, r.dir, realGh), 0o755); err != nil {
		t.Fatal(err)
	}
	pathDir = filepath.Join(r.dir, "bin")
	if err := os.MkdirAll(pathDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pathDir, "git"), fmt.Appendf(nil, sleepingGitStandIn, r.dir, realGit), 0o755); err != nil {
		t.Fatal(err)
	}
	return gh, pathDir
}

// TestRigProbe is the round plan's check of its shared assumption: a launch
// from a git-made worktree through systemd-run reaches the claude stand-in
// with the worktree as its working folder.
func TestRigProbe(t *testing.T) {
	r := newLaunchRig(t)
	wt := filepath.Join(r.clone, ".claude", "worktrees", "p1")
	git(t, r.clone, "worktree", "add", "--no-track", wt, "-b", "p1", "origin/main")

	// systemd-run's flags come first, the way the launch will build them.
	cmd := exec.Command(r.systemd, "--user", "--scope", "-q", r.claude, "--bg", "--name", "probe-p1", "a prompt\nwith \"quotes\"", "--disallowedTools", "X")
	cmd.Dir = wt
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("launch: %v\n%s", err, out)
	}
	if got, want := strings.TrimSpace(string(out)), "backgrounded · b0000001 · probe-p1"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	want, err := filepath.EvalSymlinks(wt)
	if err != nil {
		t.Fatal(err)
	}
	launches := r.launches(t)
	if len(launches) != 1 {
		t.Fatalf("launches = %d, want 1", len(launches))
	}
	if got, _ := filepath.EvalSymlinks(launches[0].Cwd); got != want {
		t.Errorf("claude's working folder = %q, want the worktree %q", got, want)
	}
	if got := launches[0].Args[len(launches[0].Args)-3]; got != "a prompt\nwith \"quotes\"" {
		t.Errorf("the prompt argument = %q", got)
	}
	scope := r.scopeLog(t)
	if !strings.Contains(scope, "flags=--user --scope -q") || !strings.Contains(scope, "cwd=") {
		t.Errorf("scope.log = %q, want the three flags and a working folder", scope)
	}
	if rows := r.readRows(t); len(rows) != 1 || rows[0]["state"] != "working" || rows[0]["name"] != "probe-p1" {
		t.Errorf("rows = %v, want one working row for probe-p1", rows)
	}
	if b, err := os.ReadFile(filepath.Join(r.jobs, "b0000001", "state.json")); err != nil || !strings.Contains(string(b), "b0000001-0000") {
		t.Errorf("job state = %q, %v; want the session id", b, err)
	}
}

// TestRigControls shows the stand-in's controls work, so the launch and
// recovery tests never edit the rig.
func TestRigControls(t *testing.T) {
	r := newLaunchRig(t)
	claude := func(args ...string) (string, error) {
		cmd := exec.Command(r.claude, args...)
		cmd.Dir = r.dir
		cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}

	if out, err := claude("--bg", "--name", "one"); err != nil || out != "backgrounded · b0000001 · one" {
		t.Fatalf("--bg = %q, %v", out, err)
	}
	r.control(t, "no-id", true)
	if out, err := claude("--bg", "--name", "two"); err != nil || out != "" {
		t.Errorf("--bg with no-id = %q, %v; want no output", out, err)
	}
	r.control(t, "no-id", false)
	r.control(t, "no-row", true)
	if _, err := claude("--bg", "--name", "three"); err != nil {
		t.Fatal(err)
	}
	r.control(t, "no-row", false)
	if got := len(r.readRows(t)); got != 2 {
		t.Errorf("rows = %d, want 2 (one and two; three had no-row)", got)
	}

	r.setList(t, 5, map[string]any{"id": "scripted"})
	var seen []string
	for range 5 {
		out, err := claude("agents", "--json", "--all")
		if err != nil {
			t.Fatal(err)
		}
		seen = append(seen, out)
	}
	if !strings.Contains(seen[4], "scripted") || strings.Contains(seen[3], "scripted") {
		t.Errorf("only the fifth read should print the scripted list: %q", seen)
	}

	r.control(t, "stop-fails", true)
	if _, err := claude("stop", "b0000001"); err == nil {
		t.Error("stop with stop-fails exited 0")
	}
	if got := r.readRows(t)[0]["state"]; got != "working" {
		t.Errorf("a failed stop left the row %v, want working", got)
	}
	r.control(t, "stop-fails", false)
	if _, err := claude("stop", "b0000001"); err != nil {
		t.Fatal(err)
	}
	if got := r.readRows(t)[0]["state"]; got != "stopped" {
		t.Errorf("a stop left the row %v, want stopped", got)
	}

	r.vanish(t, "b0000001")
	if got := len(r.readRows(t)); got != 1 {
		t.Errorf("rows after vanish = %d, want 1", got)
	}
	r.control(t, "fail-bg", true)
	if _, err := claude("--bg", "--name", "four"); err == nil {
		t.Error("--bg with fail-bg exited 0")
	}
}
