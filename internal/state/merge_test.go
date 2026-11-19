package state

import (
	"strings"
	"testing"
	"time"
)

// mergeState loads a release whose pass x is in the given state with a
// session, and the extra pass JSON fields.
func mergeState(t *testing.T, passState, extra string) *State {
	t.Helper()
	session := goodSession
	if passState == "not_started" {
		session = ""
	}
	j := sessionRelease(passState, session, "")
	if extra != "" {
		j = strings.Replace(j, `"state":"`+passState+`"`, `"state":"`+passState+`",`+extra, 1)
	}
	_, s := loadDir(t, j)
	return s
}

func pass(t *testing.T, s *State) *Pass {
	t.Helper()
	_, p, err := findPass(s, "p", "x")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

const stoppedPass = `"started":"2026-11-20T10:00:00Z","stop_reason":"scope","stops":[{"start":"2026-11-20T10:20:00Z"}]`

const readyPR = `{"number":7,"commit":"abc","verdict":"accept","inside_design":true,"ready":"2026-11-20T10:30:00Z","wait":"no_check","failures":["gate"],"question":"","merge_called":"2026-11-20T10:40:00Z"}`

func TestReportReady(t *testing.T) {
	tests := []struct {
		name, state, extra, verdict string
		refused                     bool
	}{
		{"building", "building", "", "accept", false},
		{"a second report replaces the first", "building", `"pull_request":` + readyPR, "fix", false},
		{"stopped", "stopped", stoppedPass, "accept", true},
		{"not started", "not_started", "", "accept", true},
		{"bad verdict", "building", "", "ship", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mergeState(t, tt.state, tt.extra)
			err := ReportReady("p", "x", 9, "def", tt.verdict, false)(s, at(11, 0))
			if tt.refused != isRefusal(err) || (err != nil && !tt.refused) {
				t.Fatalf("err = %v, refused want %v", err, tt.refused)
			}
			if tt.refused {
				return
			}
			pr := pass(t, s).PullRequest
			if pr.Number != 9 || pr.Commit != "def" || pr.Verdict != tt.verdict || pr.InsideDesign || *pr.Ready != at(11, 0) {
				t.Errorf("record = %+v", pr)
			}
			if pr.Wait != "" || pr.Failures != nil || pr.Question != "" || pr.MergeCalled != nil {
				t.Errorf("the report left %+v of the old one", pr)
			}
		})
	}

	s := mergeState(t, "building", `"pull_request":{"number":7,"fix_attempts":1}`)
	if err := ReportReady("p", "x", 7, "def", "accept", true)(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	if got := pass(t, s).PullRequest.FixAttempts; got != 1 {
		t.Errorf("fix attempts = %d, want 1 carried over", got)
	}

	noSession := loaded(t)
	if err := ReportReady("alpha", "one", 1, "c", "accept", true)(noSession, at(11, 0)); err == nil {
		t.Error("a report on a pass with no session passed")
	}
}

func TestMergePass(t *testing.T) {
	tests := []struct {
		name, state, extra string
		at                 time.Time
		wantFinished       time.Time
		minutes            int
		refused            bool
	}{
		{"building", "building", "", at(11, 0), at(11, 0), 60, false},
		{"stopped closes its stop", "stopped", stoppedPass, at(11, 0), at(11, 0), 20, false},
		{"merged before the stop began", "stopped", stoppedPass, at(10, 10), at(10, 20), 20, false},
		{"merged before the pass started", "building", "", at(9, 0), at(10, 0), 0, false},
		{"not started", "not_started", "", at(11, 0), time.Time{}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mergeState(t, tt.state, tt.extra)
			err := MergePass("p", "x", 7, tt.at, true)(s, at(12, 0))
			if tt.refused {
				if !isRefusal(err) {
					t.Fatalf("err = %v, want a refusal", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			p := pass(t, s)
			if p.State != Done || !p.Finished.Equal(tt.wantFinished) || *p.ActualMinutes != tt.minutes {
				t.Errorf("pass = %s finished %v minutes %d", p.State, p.Finished, *p.ActualMinutes)
			}
			if !p.PullRequest.Merged.Equal(tt.wantFinished) || !p.PullRequest.MergedOutside || p.StopReason != "" {
				t.Errorf("record = %+v", p.PullRequest)
			}
			if err := validatePass(p, map[string]*Round{"a": {}}, map[string]*Pass{}, map[string]bool{}); err != nil {
				t.Errorf("the merged pass fails its own validation: %v", err)
			}
		})
	}

	s := mergeState(t, "building", "")
	if err := MergePass("p", "x", 7, at(11, 0), false)(s, at(12, 0)); err != nil {
		t.Fatal(err)
	}
	if err := MergePass("p", "x", 7, at(11, 0), false)(s, at(12, 0)); !isRefusal(err) {
		t.Errorf("a second merge: %v, want a refusal", err)
	}
}

func TestAskMerge(t *testing.T) {
	s := mergeState(t, "building", `"pull_request":{"number":7,"commit":"abc","verdict":"accept","ready":"2026-11-20T10:30:00Z"}`)
	opts := []string{"merge", "fix again", "close"}
	for want := 1; want <= 2; want++ {
		head := "head" + string(rune('0'+want))
		if err := AskMerge("p", "x", "Pull request 7 failed.", head, opts, "close")(s, at(11, 0)); err != nil {
			t.Fatal(err)
		}
		id := "merge-x-" + string(rune('0'+want))
		rf := s.Releases[0]
		if q := rf.question(id); q == nil || q.Recommended != "close" || q.Answer != nil {
			t.Fatalf("question %s = %+v", id, q)
		}
		p := pass(t, s)
		if p.PullRequest.Question != id || p.PullRequest.AskedCommit != head || p.Questions[len(p.Questions)-1] != id || p.PullRequest.Ready == nil {
			t.Errorf("pass after ask %d = %+v", want, p)
		}
	}
	empty := mergeState(t, "building", "")
	if err := AskMerge("p", "x", "t", "abc", opts, "close")(empty, at(11, 0)); !isRefusal(err) {
		t.Errorf("ask with no report: %v, want a refusal", err)
	}
}

func TestRecordFixClearsTheQuestion(t *testing.T) {
	s := mergeState(t, "building", `"pull_request":{"number":7,"commit":"abc","verdict":"accept","fix_attempts":2,"ready":"2026-11-20T10:30:00Z"}`)
	if err := AskMerge("p", "x", "t", "abc", []string{"merge", "fix again", "close"}, "close")(s, at(10, 45)); err != nil {
		t.Fatal(err)
	}
	if err := RecordFix("p", "x", nil)(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	pr := pass(t, s).PullRequest
	if pr.Question != "" || pr.AskedCommit != "" || pr.FixAttempts != 3 || pr.Ready != nil {
		t.Errorf("pull request after the fix = %+v", pr)
	}
}

func TestRecordFix(t *testing.T) {
	s := mergeState(t, "building", `"pull_request":{"number":7,"commit":"abc","verdict":"accept","ready":"2026-11-20T10:30:00Z","wait":"check_running"}`)
	if err := RecordFix("p", "x", []string{"gate"})(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	pr := pass(t, s).PullRequest
	if pr.FixAttempts != 1 || pr.Ready != nil || pr.Wait != "" || len(pr.Failures) != 1 {
		t.Errorf("record = %+v", pr)
	}
	if err := RecordFix("p", "x", nil)(s, at(11, 0)); !isRefusal(err) {
		t.Errorf("a second fix with no new report: %v, want a refusal", err)
	}

	ev := Sequence(
		RecordFix("p", "x", nil),
		RecordLaunch("p", "pass:x", Launch{Kind: LaunchRelaunch, Replaces: "ab12", Cause: "fix", Session: bareSession("p-one", "/w/one")}),
	)
	r := mergeState(t, "building", `"pull_request":{"number":7,"commit":"abc","verdict":"fix","ready":"2026-11-20T10:30:00Z"}`)
	if err := ev(r, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	if got := pass(t, r); got.PullRequest.FixAttempts != 1 || len(got.Session.Relaunches) != 1 || got.Session.Counted() != 0 {
		t.Errorf("composed commit left %+v", got)
	}
}

func TestRecordCleanup(t *testing.T) {
	s := mergeState(t, "building", `"pull_request":{"number":7}`)
	if err := MarkMerge("p", "x")(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	if pr := pass(t, s).PullRequest; pr.MergeCalled == nil || !pr.MergeCalled.Equal(at(11, 0)) {
		t.Errorf("marker = %+v", pr)
	}
	if err := RecordCleanup("p", "x", nil)(s, at(11, 5)); !isRefusal(err) {
		t.Errorf("cleanup before the merge: %v, want a refusal", err)
	}
	if err := MergePass("p", "x", 7, at(11, 1), false)(s, at(11, 1)); err != nil {
		t.Fatal(err)
	}
	if err := RecordCleanup("p", "x", []string{"worktree left: dirty", "branch left: unmerged"})(s, at(11, 5)); err != nil {
		t.Fatal(err)
	}
	pr := pass(t, s).PullRequest
	if !pr.Cleaned.Equal(at(11, 5)) || pr.CleanupNote != "worktree left: dirty; branch left: unmerged" {
		t.Errorf("record = %+v", pr)
	}
}

func TestPullRequestFields(t *testing.T) {
	tests := []struct {
		name, pr, want string
	}{
		{"a ready time alone, as old fixtures hold", `{"number":9,"ready":"2026-11-20T10:30:00Z"}`, ""},
		{"all fields", readyPR, ""},
		{"bad verdict", `{"number":9,"verdict":"ship"}`, "verdict"},
		{"bad wait", `{"number":9,"wait":"soon"}`, "wait"},
		{"merged while building", `{"number":9,"merged":"2026-11-20T10:30:00Z"}`, "merged"},
		{"question that does not exist", `{"number":9,"question":"merge-x-1"}`, "question"},
		{"unknown key", `{"number":9,"readyy":1}`, "unknown field"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := strings.Replace(sessionRelease("building", goodSession, ""), `"state":"building"`, `"state":"building","pull_request":`+tt.pr, 1)
			_, err := DecodeRelease([]byte(j))
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRecordCheckWritesPRWait(t *testing.T) {
	s := mergeState(t, "building", `"pull_request":{"number":7,"wait":"no_check"}`)
	d := Decision{Project: "p", Record: "pass:x", Session: "p-one", Decision: "none"}
	empty := ""
	if err := RecordCheck([]Recorded{{Decision: d, PRWait: &empty}})(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	if got := pass(t, s).PullRequest.Wait; got != "" {
		t.Errorf("wait = %q, want it cleared", got)
	}
	w := WaitCheckRunning
	if err := RecordCheck([]Recorded{{Decision: d, PRWait: &w}})(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	if got := pass(t, s).PullRequest.Wait; got != w {
		t.Errorf("wait = %q", got)
	}
}
