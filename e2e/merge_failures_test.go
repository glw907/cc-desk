//go:build unix

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// failingMerge is a gh stand-in that fails `pr merge` while the rig folder
// holds a merge-fails file, and otherwise hands every call to the check-merge
// stand-in beside it.
const failingMerge = `#!/usr/bin/env python3
import json
import os
import sys

d = os.path.dirname(os.path.abspath(__file__))
a = sys.argv[1:]
if a[:2] == ["pr", "merge"] and os.path.exists(os.path.join(d, "merge-fails")):
    with open(os.path.join(d, "gh-calls.log"), "a") as f:
        f.write(json.dumps(a) + "\n")
    sys.stderr.write("gh: pull request is not mergeable\n")
    sys.exit(1)
os.execv(sys.executable, [sys.executable, os.path.join(d, "gh.py")] + a)
`

// mergePullField reads a number field of the pass's pull request record, 0
// when absent.
func (m mergeRig) mergePullField(t *testing.T, pass, field string) int {
	t.Helper()
	n, _ := m.pullOf(t, pass)[field].(float64)
	return int(n)
}

// questionText is the text of the release's question with the id.
func (m mergeRig) questionText(t *testing.T, id string) any {
	t.Helper()
	for _, q := range m.release(t)["questions"].([]any) {
		if qm := q.(map[string]any); qm["id"] == id {
			return qm["text"]
		}
	}
	t.Fatalf("question %s is not in the release", id)
	return nil
}

func TestMergeFailures(t *testing.T) {
	m := newMergeRig(t, "mf")
	wrapper := filepath.Join(m.dir, "gh-failing.py")
	if err := os.WriteFile(wrapper, []byte(failingMerge), 0o755); err != nil {
		t.Fatal(err)
	}
	m.gh = wrapper
	head := m.push(t, "mf", "mf.txt")
	m.report(t, "mf", 1, head, "inside")
	m.answer(t, "mf", 1, head, passing)
	m.rowsDone(t)
	m.write(t, "merge-fails", "")

	// failRuns runs the check three times against the failing merge call.
	// Runs 1 and 2 log the count and exit 0; run 3 logs it and asks.
	failRuns := func(t *testing.T, question string) {
		t.Helper()
		for n := 1; n <= 3; n++ {
			requireRun(t, m.check(t, mergeClock, ""))
			want := fmt.Sprintf("merge call failed (%d of 3): gh: pull request is not mergeable", n)
			if got := lastOutcome(t, m.state); got != want {
				t.Fatalf("run %d outcome = %q, want %q", n, got, want)
			}
			pr := m.pullOf(t, "mf")
			if n < 3 {
				if got := m.mergePullField(t, "mf", "merge_failures"); got != n || pr["question"] == question {
					t.Fatalf("run %d: failures %d, question %v, want %d and none", n, got, pr["question"], n)
				}
				continue
			}
			if pr["question"] != question || m.mergePullField(t, "mf", "merge_failures") != 0 {
				t.Fatalf("run 3: pull request = %v, want question %s and the count cleared", pr, question)
			}
			text := fmt.Sprint(m.questionText(t, question))
			if !strings.Contains(text, "merge call") || !strings.Contains(text, "not mergeable") {
				t.Errorf("question text = %q", text)
			}
		}
	}

	failRuns(t, "merge-mf-1")
	if got := len(m.merges(t)); got != 3 {
		t.Fatalf("%d merge calls, want 3", got)
	}
	// The question holds the pull request: a fourth run makes no call.
	requireRun(t, m.check(t, mergeClock, ""))
	if got := len(m.merges(t)); got != 3 {
		t.Fatalf("%d merge calls while the question is open, want 3", got)
	}

	// A "merge" answer gets three more calls.
	m.answerQuestion(t, "merge-mf-1", "merge", "Try again.")
	failRuns(t, "merge-mf-2")
	if got := len(m.merges(t)); got != 6 {
		t.Fatalf("%d merge calls after the answer, want 6", got)
	}
}
