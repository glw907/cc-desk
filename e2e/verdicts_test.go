//go:build unix

package e2e

import (
	"slices"
	"strings"
	"testing"
)

func TestVerdicts(t *testing.T) {
	passes := []string{"acc", "fx", "esc"}
	verdicts := map[string]string{"acc": "accept", "fx": "fix", "esc": "escalate"}
	want := map[string]string{"acc": "merge", "fx": "fix", "esc": "merge"}
	m := newMergeRig(t, passes...)
	heads := map[string]string{}
	for i, p := range passes {
		heads[p] = m.push(t, p, p+".txt")
		requireRun(t, m.reportAs(t, p, i+1, heads[p], "inside", verdicts[p], 0))
		m.answer(t, p, i+1, heads[p], passing)
	}
	m.rowsDone(t)
	out := m.out(t, m.check(t, mergeClock, ""))
	for _, p := range passes {
		if got := out["pass:"+p].Decision; got != want[p] {
			t.Errorf("%s (%s): decision %q, want %q", p, verdicts[p], got, want[p])
		}
	}
	if n := len(m.merges(t)); n != 2 {
		t.Errorf("merge calls = %d, want 2 (accept and escalate)", n)
	}
}

func TestCallsForGeoff(t *testing.T) {
	m := answersRig(t, "p", "f")
	head := m.push(t, "p", "p.txt")
	requireRun(t, m.reportAs(t, "p", 1, head, "inside", "escalate", 1))
	m.answer(t, "p", 1, head, passing)
	fhead := m.push(t, "f", "f.txt")
	requireRun(t, m.reportAs(t, "f", 2, fhead, "inside", "fix", 2))
	m.answer(t, "f", 2, fhead, passing)
	m.rowsDone(t)

	out := m.out(t, m.check(t, mergeClock, ""))
	if d := out["pass:p"]; d.Decision != "ask" || d.Reason != "calls_for_geoff" {
		t.Fatalf("escalate with calls: %+v, want ask for calls_for_geoff", d)
	}
	if d := out["pass:f"]; d.Decision != "fix" {
		t.Errorf("fix with calls: %+v, want fix", d)
	}
	if n := len(m.merges(t)); n != 0 {
		t.Fatalf("%d merge calls before the answer, want none", n)
	}
	pr := m.pullOf(t, "p")
	if pr["question"] != "merge-p-1" || pr["calls"] != float64(1) || pr["asked_commit"] != head {
		t.Fatalf("pull request = %v, want merge-p-1 asked at %s with calls 1", pr, head)
	}
	for _, q := range m.release(t)["questions"].([]any) {
		qm := q.(map[string]any)
		if qm["id"] == "merge-p-1" && qm["recommended"] != "merge" {
			t.Errorf("question = %v, want merge recommended", qm)
		}
	}

	m.answerQuestion(t, "merge-p-1", "merge", "Merge it.")
	d := m.out(t, m.check(t, mergeClock, ""))["pass:p"]
	if d.Decision != "merge" {
		t.Fatalf("after the answer: %+v, want merge with no second question", d)
	}
	merges := m.merges(t)
	if len(merges) != 1 || !slices.Contains(merges[0], head) {
		t.Errorf("merge calls = %v, want one pinned to %s", merges, head)
	}
	if got := m.pullOf(t, "p")["question"]; got != "merge-p-1" {
		t.Errorf("question = %v, want still merge-p-1", got)
	}

	got := m.reportAs(t, "f", 2, fhead, "inside", "review", 0)
	for _, w := range []string{"accept", "fix", "escalate"} {
		if got.code == 0 || !strings.Contains(got.stderr, w) {
			t.Errorf("--verdict review: exit %d, stderr %q, want a refusal naming %s", got.code, got.stderr, w)
		}
	}
}
