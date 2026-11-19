package state

import "testing"

func TestMergeFailures(t *testing.T) {
	const pr = `"pull_request":{"number":7,"commit":"abc","verdict":"accept","ready":"2026-11-20T10:30:00Z"}`
	failures := func(s *State) int { return pass(t, s).PullRequest.MergeFailures }
	count := func(t *testing.T, s *State, n int) {
		t.Helper()
		for range n {
			if err := RecordMergeFailure("p", "x")(s, at(11, 0)); err != nil {
				t.Fatal(err)
			}
		}
	}

	t.Run("a failure raises the count", func(t *testing.T) {
		s := mergeState(t, "building", pr)
		count(t, s, 2)
		if got := failures(s); got != 2 {
			t.Errorf("count = %d, want 2", got)
		}
		if err := RecordMergeFailure("p", "x")(mergeState(t, "building", ""), at(11, 0)); !isRefusal(err) {
			t.Errorf("no pull request record: %v, want a refusal", err)
		}
	})
	t.Run("MergePass zeroes it", func(t *testing.T) {
		s := mergeState(t, "building", pr)
		count(t, s, 2)
		if err := MergePass("p", "x", 7, at(11, 0), false)(s, at(12, 0)); err != nil {
			t.Fatal(err)
		}
		if got := failures(s); got != 0 {
			t.Errorf("count = %d, want 0", got)
		}
	})
	t.Run("AskMerge resets it", func(t *testing.T) {
		s := mergeState(t, "building", pr)
		count(t, s, 3)
		if err := AskMerge("p", "x", "t", "abc", []string{"merge", "close"}, "close")(s, at(11, 0)); err != nil {
			t.Fatal(err)
		}
		if got := failures(s); got != 0 {
			t.Errorf("count = %d, want 0", got)
		}
	})
	t.Run("a report with the same number keeps it", func(t *testing.T) {
		s := mergeState(t, "building", pr)
		count(t, s, 2)
		if err := ReportReady("p", "x", 7, "def", VerdictAccept, true, 0)(s, at(11, 0)); err != nil {
			t.Fatal(err)
		}
		if got := failures(s); got != 2 {
			t.Errorf("count = %d, want 2", got)
		}
	})
	t.Run("a new pull request starts at 0", func(t *testing.T) {
		s := mergeState(t, "building", pr)
		count(t, s, 2)
		if err := ReportReady("p", "x", 8, "def", VerdictAccept, true, 0)(s, at(11, 0)); err != nil {
			t.Fatal(err)
		}
		if got := failures(s); got != 0 {
			t.Errorf("count = %d, want 0", got)
		}
	})
}
