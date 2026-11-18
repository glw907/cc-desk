//go:build unix

package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// mergeClock is the clock of the check that decides a merge: an hour after
// workedClock, when every pass in a merge test started.
const mergeClock = "2026-11-21T00:00:00Z"

// mergeRig is a launch rig with a gh stand-in beside the claude one. The
// stand-in is testdata/check-merge/gh.py, copied into the rig folder.
type mergeRig struct {
	launchRig
	gh string
}

// newMergeRig makes the rig and a release with the named building passes,
// then runs one check at workedClock so each pass has a session in its own
// worktree and a branch of its own name.
func newMergeRig(t *testing.T, passes ...string) mergeRig {
	t.Helper()
	m := mergeRig{launchRig: newLaunchRig(t)}
	src, err := os.ReadFile(filepath.Join("testdata", "check-merge", "gh.py"))
	if err != nil {
		t.Fatal(err)
	}
	m.gh = filepath.Join(m.dir, "gh.py")
	if err := os.WriteFile(m.gh, src, 0o755); err != nil {
		t.Fatal(err)
	}
	m.write(t, "remote", m.remote)
	if err := os.MkdirAll(filepath.Join(m.dir, "answers"), 0o755); err != nil {
		t.Fatal(err)
	}
	m.addRelease(t, true, passes...)
	m.building(t, passes...)
	first := m.check(t, workedClock, "")
	requireRun(t, first)
	if got := len(m.launches(t)); got != len(passes) {
		t.Fatalf("%d launches, want %d; check printed:\n%s%s", got, len(passes), first.stdout, first.stderr)
	}
	return m
}

// building writes the named passes into the release file as building since
// workedClock. The start verb holds a third building pass back, and the
// merge tests need up to six.
func (m mergeRig) building(t *testing.T, passes ...string) {
	t.Helper()
	file := filepath.Join(m.state, "releases", "proj.json")
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var rel map[string]any
	if err := json.Unmarshal(b, &rel); err != nil {
		t.Fatal(err)
	}
	for _, p := range rel["passes"].([]any) {
		pass := p.(map[string]any)
		if slices.Contains(passes, pass["name"].(string)) {
			pass["state"], pass["started"] = "building", workedClock
		}
	}
	if b, err = json.MarshalIndent(rel, "", "  "); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// check runs the check with every stand-in flag, including the gh one. With
// a trace file, git in the binary's environment writes its trace there.
func (m mergeRig) check(t *testing.T, now, trace string) result {
	t.Helper()
	args := append(m.args(t, "check"), "--gh-command", m.gh)
	var env []string
	if trace != "" {
		env = append(env, "GIT_TRACE="+trace)
	}
	return runEnv(t, m.state, now, env, args...)
}

// runEnv is run with extra environment for the binary alone.
func runEnv(t *testing.T, dir, now string, env []string, args ...string) result {
	t.Helper()
	full := append([]string{"--state-dir", dir, "--now", now, "--usage-command", usageStandIn(t)}, args...)
	cmd := exec.Command(binary, full...)
	cmd.Env = append(append(os.Environ(), "PATH=/usr/bin:/bin"), env...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	code := 0
	if err := cmd.Run(); err != nil {
		ee, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			t.Fatalf("run %v: %v", args, err)
		}
		code = ee.ExitCode()
	}
	return result{out.String(), errOut.String(), code}
}

// worktree is the pass's worktree folder.
func (m mergeRig) worktree(pass string) string {
	return filepath.Join(m.clone, ".claude", "worktrees", pass)
}

// push commits a file in the pass's worktree and pushes the branch, as the
// session would, and returns the new head.
func (m mergeRig) push(t *testing.T, pass, file string) string {
	t.Helper()
	commitFile(t, m.worktree(pass), file, pass+"\n", "Add "+file)
	git(t, m.worktree(pass), "push", "origin", pass)
	return git(t, m.worktree(pass), "rev-parse", "HEAD")
}

// check kinds for the rollup of a pull request's answer.
var (
	passing = map[string]any{"name": "gate", "status": "COMPLETED", "conclusion": "SUCCESS"}
	failing = map[string]any{"name": "gate", "status": "COMPLETED", "conclusion": "FAILURE"}
	running = map[string]any{"name": "gate", "status": "IN_PROGRESS", "conclusion": ""}
)

// answer sets what `pr list --head <pass>` prints: one open pull request
// with the head and the rollup's checks.
func (m mergeRig) answer(t *testing.T, pass string, number int, head string, checks ...map[string]any) {
	t.Helper()
	if checks == nil {
		checks = []map[string]any{}
	}
	m.setAnswer(t, pass, map[string]any{
		"number": number, "state": "OPEN", "headRefOid": head,
		"mergedAt": nil, "mergeCommit": nil, "statusCheckRollup": checks,
	})
}

// setAnswer writes the stand-in's answer file for a pass's branch.
func (m mergeRig) setAnswer(t *testing.T, pass string, pull map[string]any) {
	t.Helper()
	b, err := json.Marshal([]any{pull})
	if err != nil {
		t.Fatal(err)
	}
	m.write(t, filepath.Join("answers", pass+".json"), string(b))
}

// report records the session's ready report for the pass.
func (m mergeRig) report(t *testing.T, pass string, number int, commit, design string) {
	t.Helper()
	requireRun(t, m.do(t, workedClock, "pass", "ready", "proj", pass,
		"--pr", strconv.Itoa(number), "--commit", commit, "--verdict", "accept", "--design", design))
}

// rowsDone sets every row the stand-in holds to done and idle, so the row
// rules leave a waiting pass alone.
func (m mergeRig) rowsDone(t *testing.T) {
	t.Helper()
	rows := m.readRows(t)
	for _, r := range rows {
		r["state"], r["status"] = "done", "idle"
	}
	m.setRows(t, rows...)
}

// ghCalls returns the stand-in gh's calls, each an argument list.
func (m mergeRig) ghCalls(t *testing.T) [][]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(m.dir, "gh-calls.log"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out [][]string
	for line := range strings.SplitSeq(strings.TrimSpace(string(b)), "\n") {
		var a []string
		if err := json.Unmarshal([]byte(line), &a); err != nil {
			t.Fatalf("gh-calls.log line %q: %v", line, err)
		}
		out = append(out, a)
	}
	return out
}

// merges returns the stand-in gh's `pr merge` calls.
func (m mergeRig) merges(t *testing.T) [][]string {
	t.Helper()
	var out [][]string
	for _, a := range m.ghCalls(t) {
		if len(a) > 1 && a[0] == "pr" && a[1] == "merge" {
			out = append(out, a)
		}
	}
	return out
}

// mergeOutside merges the branch into main on the remote with git, as
// another hand would, and returns the merge commit.
func (m mergeRig) mergeOutside(t *testing.T, branch string) string {
	t.Helper()
	other := filepath.Join(t.TempDir(), "other")
	git(t, m.dir, "clone", m.remote, other)
	git(t, other, "merge", "--no-ff", "origin/"+branch, "-m", "Merge "+branch)
	git(t, other, "push", "origin", "main")
	return git(t, other, "rev-parse", "HEAD")
}

// out parses the decision lines of a check run.
func (m mergeRig) out(t *testing.T, got result) map[string]decision {
	t.Helper()
	requireRun(t, got)
	return decisions(t, got.stdout)
}
