//go:build unix

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// pullOf is the pass's pull_request object from the release file.
func (m mergeRig) pullOf(t *testing.T, pass string) map[string]any {
	t.Helper()
	pr, _ := m.passRecord(t, pass)["pull_request"].(map[string]any)
	if pr == nil {
		t.Fatalf("pass %q has no pull request: %v", pass, m.passRecord(t, pass))
	}
	return pr
}

// traceTokens returns the words of every command git ran for the binary, from
// the trace's built-in and run_command lines.
func traceTokens(t *testing.T, file string) []string {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for line := range strings.SplitSeq(string(b), "\n") {
		for _, marker := range []string{"trace: built-in:", "trace: run_command:"} {
			if _, rest, ok := strings.Cut(line, marker); ok {
				for tok := range strings.FieldsSeq(rest) {
					out = append(out, strings.Trim(tok, "'"))
				}
			}
		}
	}
	return out
}

func TestMergeSix(t *testing.T) {
	passes := []string{"clean", "failed", "outside", "nocheck", "running", "moved"}
	m := newMergeRig(t, passes...)
	heads := map[string]string{}
	for _, p := range passes {
		heads[p] = m.push(t, p, p+".txt")
	}
	reports := []struct {
		pass, design string
		checks       []map[string]any
	}{
		{"clean", "inside", []map[string]any{passing}},
		{"failed", "inside", []map[string]any{failing}},
		{"outside", "outside", []map[string]any{passing}},
		{"nocheck", "inside", nil},
		{"running", "inside", []map[string]any{running}},
		{"moved", "inside", []map[string]any{passing}},
	}
	for i, r := range reports {
		m.report(t, r.pass, i+1, heads[r.pass], r.design)
		answerHead := heads[r.pass]
		if r.pass == "moved" {
			answerHead = m.push(t, "moved", "moved-again.txt")
		}
		m.answer(t, r.pass, i+1, answerHead, r.checks...)
	}
	m.rowsDone(t)

	trace := filepath.Join(t.TempDir(), "trace")
	m.out(t, m.check(t, mergeClock, trace))

	merges := m.merges(t)
	if len(merges) != 1 {
		t.Fatalf("pr merge calls = %v, want exactly one", merges)
	}
	call := strings.Join(merges[0], " ")
	if !strings.HasPrefix(call, "pr merge 1 ") || !slices.Contains(merges[0], "--merge") || !strings.HasSuffix(call, "--match-head-commit "+heads["clean"]) {
		t.Errorf("merge call = %q, want pull request 1 with --merge and the head %s", call, heads["clean"])
	}
	for _, bad := range []string{"--admin", "-d", "--delete-branch", "--auto"} {
		if slices.Contains(merges[0], bad) {
			t.Errorf("merge call holds %s: %q", bad, call)
		}
	}

	clean := m.passRecord(t, "clean")
	pr := m.pullOf(t, "clean")
	if clean["state"] != "done" || clean["finished"] != mergeClock || pr["merged"] != mergeClock || pr["cleaned"] == nil {
		t.Errorf("clean pass = %v, want done, finished and merged at %s, and cleaned", clean, mergeClock)
	}
	if _, err := os.Stat(m.worktree("clean")); !os.IsNotExist(err) {
		t.Errorf("the clean worktree is still there (stat error %v)", err)
	}
	if got := git(t, m.clone, "branch", "--list", "clean"); got != "" {
		t.Errorf("branch clean is still there: %q", got)
	}
	if got := git(t, m.clone, "log", "--merges", "--format=%s", "main"); got != "Merge pull request 1" {
		t.Errorf("main's merge commits = %q, want the one for pull request 1", got)
	}
	tokens := traceTokens(t, trace)
	for _, want := range []string{"--ff-only", "-d", "remove"} {
		if !slices.Contains(tokens, want) {
			t.Errorf("the trace holds no %q, so it did not see the cleanup", want)
		}
	}
	for _, tok := range tokens {
		if tok == "-D" || tok == "-f" || tok == "--force" {
			t.Errorf("git ran with %s", tok)
		}
	}

	failed := m.sessionOf(t, "failed")
	rel, _ := failed["relaunches"].([]any)
	if len(rel) != 1 {
		t.Fatalf("failed pass relaunches = %v, want one", failed["relaunches"])
	}
	if r := rel[0].(map[string]any); r["cause"] != "fix" || r["counted"] == true {
		t.Errorf("relaunch = %v, want cause fix, uncounted", r)
	}
	if got := len(m.launches(t)); got != len(passes)+1 {
		t.Errorf("launches = %d, want %d: one per pass and the fix", got, len(passes)+1)
	}
	if n := m.promptsWith(t, "Pull request 2 came back at commit "+heads["failed"]+": gate"); n != 1 {
		t.Errorf("launches carrying the fix note = %d, want 1", n)
	}

	if q := m.pullOf(t, "outside")["question"]; q != "merge-outside-1" {
		t.Errorf("outside pass question = %v, want merge-outside-1", q)
	}
	if s := readEst(t, m.state, mergeClock); len(s.WaitsForGeoff) != 1 || s.WaitsForGeoff[0].Questions != 1 {
		t.Errorf("waits_for_geoff = %+v, want one wait with one question", s.WaitsForGeoff)
	}
	for pass, wait := range map[string]string{"nocheck": "no_check", "running": "check_running", "moved": "commit_moved"} {
		if got := m.pullOf(t, pass)["wait"]; got != wait {
			t.Errorf("%s wait = %v, want %s", pass, got, wait)
		}
		if m.passRecord(t, pass)["state"] != "building" {
			t.Errorf("%s left building: %v", pass, m.passRecord(t, pass)["state"])
		}
	}
}

// promptsWith counts the launches whose prompt holds the text.
func (m mergeRig) promptsWith(t *testing.T, text string) int {
	t.Helper()
	n := 0
	for _, c := range m.launches(t) {
		if slices.ContainsFunc(c.Args, func(a string) bool { return strings.Contains(a, text) }) {
			n++
		}
	}
	return n
}

func TestMergeFixLimit(t *testing.T) {
	m := newMergeRig(t, "fx")
	sent := 0
	send := func() string {
		sent++
		head := m.push(t, "fx", "fix"+strconv.Itoa(sent)+".txt")
		m.report(t, "fx", 1, head, "inside")
		m.answer(t, "fx", 1, head, failing)
		m.rowsDone(t)
		return head
	}

	first := send()
	m.out(t, m.check(t, workedClock, ""))
	if got := len(m.launches(t)); got != 2 {
		t.Fatalf("launches after the first failure = %d, want 2", got)
	}
	if m.promptsWith(t, "Pull request 1 came back at commit "+first) != 1 {
		t.Error("the first relaunch lacks the fix note")
	}

	m.rowsDone(t)
	m.out(t, m.check(t, workedClock, ""))
	if got, merges := len(m.launches(t)), len(m.merges(t)); got != 2 || merges != 0 {
		t.Errorf("a check with no new report: %d launches, %d merge calls, want 2 and 0", got, merges)
	}

	send()
	m.out(t, m.check(t, workedClock, ""))
	if got := len(m.launches(t)); got != 3 {
		t.Fatalf("launches after the second failure = %d, want 3", got)
	}

	send()
	d := m.out(t, m.check(t, workedClock, ""))["pass:fx"]
	if d.Decision != "ask" || d.Reason != "fix_limit" {
		t.Errorf("third failure decision = %+v, want ask for fix_limit", d)
	}
	if got, merges := len(m.launches(t)), len(m.merges(t)); got != 3 || merges != 0 {
		t.Errorf("after the third failure: %d launches, %d merge calls, want 3 and 0", got, merges)
	}
	if q := m.pullOf(t, "fx")["question"]; q != "merge-fx-1" {
		t.Errorf("question = %v, want merge-fx-1", q)
	}
	if got := m.pullOf(t, "fx")["fix_attempts"]; got != float64(2) {
		t.Errorf("fix_attempts = %v, want 2", got)
	}
}

func TestMergeNoCheck(t *testing.T) {
	m := newMergeRig(t, "nc")
	head := m.push(t, "nc", "nc.txt")
	m.report(t, "nc", 4, head, "inside")
	m.answer(t, "nc", 4, head)
	m.rowsDone(t)

	for _, tc := range []struct {
		now  string
		want bool
	}{{"2026-11-20T23:29:00Z", false}, {"2026-11-20T23:31:00Z", true}} {
		m.out(t, m.check(t, tc.now, ""))
		if got := m.pullOf(t, "nc")["wait"]; got != "no_check" {
			t.Fatalf("at %s wait = %v, want no_check", tc.now, got)
		}
		var snap struct {
			NeedsYouNow []struct{ Reason, Cause, Todo string } `json:"needs_you_now"`
			NextSession struct{ State string }                 `json:"next_session"`
		}
		if err := json.Unmarshal([]byte(readFileIn(t, m.state, "snapshot.json")), &snap); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, n := range snap.NeedsYouNow {
			if n.Reason == "no_check_registered" {
				found = true
				if n.Cause != "4" || n.Todo != "See why no check has started on pull request 4" {
					t.Errorf("entry = %+v", n)
				}
			}
		}
		if found != tc.want {
			t.Errorf("at %s: no_check_registered entry = %v, want %v", tc.now, found, tc.want)
		}
		if tc.want && snap.NextSession.State != "needs_you_now" {
			t.Errorf("at %s next_session.state = %s, want needs_you_now", tc.now, snap.NextSession.State)
		}
		if !tc.want && snap.NextSession.State == "needs_you_now" {
			t.Errorf("at %s next_session.state = needs_you_now, want something else", tc.now)
		}
	}
	if got := len(m.merges(t)); got != 0 {
		t.Errorf("merge calls = %d, want none", got)
	}
}

// outsideMerged sets up a pass whose branch another hand merged at 23:40, an
// hour before mergeClock, and returns its merge commit.
func (m mergeRig) outsideMerged(t *testing.T, pass string) {
	t.Helper()
	head := m.push(t, pass, pass+".txt")
	oid := m.mergeOutside(t, pass)
	m.setAnswer(t, pass, map[string]any{
		"number": 9, "state": "MERGED", "headRefOid": head,
		"mergedAt": "2026-11-20T23:40:00Z", "mergeCommit": map[string]any{"oid": oid}, "statusCheckRollup": []any{},
	})
	m.rowsDone(t)
}

func TestMergeOutside(t *testing.T) {
	m := newMergeRig(t, "oo")
	m.outsideMerged(t, "oo")
	d := m.out(t, m.check(t, mergeClock, ""))["pass:oo"]
	if d.Decision != "merged" {
		t.Errorf("decision = %+v, want merged", d)
	}
	pass, pr := m.passRecord(t, "oo"), m.pullOf(t, "oo")
	if pass["state"] != "done" || pass["finished"] != "2026-11-20T23:40:00Z" || pr["merged"] != "2026-11-20T23:40:00Z" ||
		pr["merged_outside"] != true || pr["cleaned"] == nil {
		t.Errorf("pass = %v, pull request = %v; want done at 23:40, merged outside, and cleaned", pass, pr)
	}
	if got := len(m.merges(t)); got != 0 {
		t.Errorf("merge calls = %d, want none", got)
	}
	if _, err := os.Stat(m.worktree("oo")); !os.IsNotExist(err) {
		t.Errorf("the worktree is still there (stat error %v)", err)
	}
	if got := git(t, m.clone, "branch", "--list", "oo"); got != "" {
		t.Errorf("branch oo is still there: %q", got)
	}
}

func TestMergeDirty(t *testing.T) {
	m := newMergeRig(t, "dd")
	m.outsideMerged(t, "dd")
	if err := os.WriteFile(filepath.Join(m.worktree("dd"), "scratch.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m.out(t, m.check(t, mergeClock, ""))
	if _, err := os.Stat(m.worktree("dd")); err != nil {
		t.Errorf("the dirty worktree was removed: %v", err)
	}
	if got := git(t, m.clone, "branch", "--list", "dd"); got == "" {
		t.Error("the branch was removed while its worktree stayed")
	}
	note, _ := m.pullOf(t, "dd")["cleanup_note"].(string)
	if !strings.Contains(note, "worktree left") {
		t.Errorf("cleanup note = %q, want it to say the worktree was left", note)
	}
	var snap struct {
		Sessions []struct{ Name, Cleanup string } `json:"sessions"`
	}
	if err := json.Unmarshal([]byte(readFileIn(t, m.state, "snapshot.json")), &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.Sessions) != 1 || snap.Sessions[0].Cleanup != note {
		t.Errorf("snapshot sessions = %+v, want the cleanup note %q", snap.Sessions, note)
	}
	if got := m.pullOf(t, "dd")["cleaned"]; got == nil {
		t.Error("cleanup did not record itself, so the next check would repeat it")
	}
}

func TestMergeUnreachable(t *testing.T) {
	m := newMergeRig(t, "un")
	head := m.push(t, "un", "un.txt")
	m.report(t, "un", 5, head, "inside")
	m.answer(t, "un", 5, head, passing)
	m.control(t, "down", true)
	m.rowsDone(t)

	d := m.out(t, m.check(t, mergeClock, ""))["pass:un"]
	if d.Decision != "none" {
		t.Errorf("decision = %+v, want the row rules' none", d)
	}
	if got := len(m.merges(t)); got != 0 {
		t.Errorf("merge calls = %d, want none", got)
	}
	if got := m.pullOf(t, "un")["wait"]; got != "unreachable" {
		t.Errorf("wait = %v, want unreachable", got)
	}
	if m.passRecord(t, "un")["state"] != "building" {
		t.Errorf("pass state = %v, want building", m.passRecord(t, "un")["state"])
	}
}

func TestPassReady(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a", "b")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	requireRun(t, r.run(t, workedClock, "check"))

	flags := []string{"--pr", "7", "--commit", "abc123", "--verdict", "accept", "--design", "inside"}
	ready := func(pass string, flags []string) result {
		return r.do(t, workedClock, append([]string{"pass", "ready", "proj", pass}, flags...)...)
	}

	for i := 0; i < len(flags); i += 2 {
		rest := append(append([]string{}, flags[:i]...), flags[i+2:]...)
		got := ready("a", rest)
		if got.code == 0 || !strings.Contains(got.stderr, "required flag") || !strings.Contains(got.stderr, flags[i][2:]) {
			t.Errorf("without %s: exit %d, stderr %q, want cobra's required-flag error", flags[i], got.code, got.stderr)
		}
	}
	if pr := r.passRecord(t, "a")["pull_request"]; pr != nil {
		t.Errorf("refused calls left a pull request: %v", pr)
	}

	requireRun(t, ready("a", flags))
	pr, ok := r.passRecord(t, "a")["pull_request"].(map[string]any)
	if !ok {
		t.Fatalf("no pull request on the pass: %v", r.passRecord(t, "a"))
	}
	if pr["number"] != float64(7) || pr["commit"] != "abc123" || pr["verdict"] != "accept" || pr["inside_design"] != true || pr["ready"] == nil {
		t.Errorf("pull request = %v, want number 7, commit abc123, verdict accept, inside_design, and ready", pr)
	}

	// A pass that is not building is refused, and nothing is written.
	got := ready("b", flags)
	if got.code != 2 || !strings.Contains(got.stderr, "not_started") {
		t.Errorf("a pass not building: exit %d, stderr %q, want a refusal", got.code, got.stderr)
	}
	if pr := r.passRecord(t, "b")["pull_request"]; pr != nil {
		t.Errorf("the refused report left a pull request: %v", pr)
	}
}
