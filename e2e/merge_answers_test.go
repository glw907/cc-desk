//go:build unix

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// answersRig is a merge rig whose gh stand-in also closes pull requests.
func answersRig(t *testing.T, passes ...string) mergeRig {
	t.Helper()
	return newMergeRigFrom(t, "check-merge-answers", passes...)
}

// askOutside pushes the pass, reports pull request 1 outside the approved
// design with passing checks, and runs the check that asks the first merge
// question. It returns the head the question asks about.
func (m mergeRig) askOutside(t *testing.T, pass string) string {
	t.Helper()
	head := m.push(t, pass, pass+".txt")
	m.report(t, pass, 1, head, "outside")
	m.answer(t, pass, 1, head, passing)
	m.rowsDone(t)
	d := m.out(t, m.check(t, mergeClock, ""))["pass:"+pass]
	if d.Decision != "ask" || d.Reason != "outside_design" {
		t.Fatalf("decision = %+v, want ask for outside_design", d)
	}
	if got := m.pullOf(t, pass)["asked_commit"]; got != head {
		t.Fatalf("asked_commit = %v, want %s", got, head)
	}
	return head
}

// answerQuestion writes a ruling that answers the merge question with the
// option, and the answer on the question, as `ruling record` would. It
// returns the ruling's id.
func (m mergeRig) answerQuestion(t *testing.T, question, option, words string) string {
	t.Helper()
	var prev string
	if b, err := os.ReadFile(filepath.Join(m.state, "rulings.jsonl")); err == nil {
		prev = string(b)
	}
	id := fmt.Sprintf("r-%04d", strings.Count(prev, "\n")+1)
	line, err := json.Marshal(map[string]any{
		"format": 1, "id": id, "time": "2026-11-20T23:30:00Z", "desk": "d", "words": words,
		"project": "proj", "round": "r1", "question": question, "option": option,
	})
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, m.state, "rulings.jsonl", prev+string(line)+"\n")
	rel := m.release(t)
	found := false
	for _, q := range rel["questions"].([]any) {
		if qm := q.(map[string]any); qm["id"] == question {
			qm["answer"], found = map[string]any{"ruling": id}, true
		}
	}
	if !found {
		t.Fatalf("question %s is not in the release", question)
	}
	body, err := json.MarshalIndent(rel, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, m.state, "releases/proj.json", string(body))
	return id
}

func TestMergeAnswerMerges(t *testing.T) {
	m := answersRig(t, "p")
	head := m.askOutside(t, "p")
	m.answerQuestion(t, "merge-p-1", "merge", "Merge it.")

	d := m.out(t, m.check(t, mergeClock, ""))["pass:p"]
	if d.Decision != "merge" {
		t.Fatalf("decision = %+v, want merge, with no second question", d)
	}
	merges := m.merges(t)
	if len(merges) != 1 || !slices.Contains(merges[0], "--match-head-commit") || !slices.Contains(merges[0], head) {
		t.Fatalf("merge calls = %v, want one pinned to %s", merges, head)
	}
	if got := m.passRecord(t, "p")["state"]; got != "done" {
		t.Errorf("pass state = %v, want done", got)
	}
}

func TestMergeAnswerHeadMoved(t *testing.T) {
	m := answersRig(t, "p")
	m.askOutside(t, "p")
	m.answerQuestion(t, "merge-p-1", "merge", "Merge it.")

	moved := m.push(t, "p", "more.txt")
	m.answer(t, "p", 1, moved, passing)
	d := m.out(t, m.check(t, mergeClock, ""))["pass:p"]
	if d.Decision != "ask" || d.Reason != "head_moved" {
		t.Fatalf("decision = %+v, want ask for head_moved", d)
	}
	if n := len(m.merges(t)); n != 0 {
		t.Fatalf("%d merge calls after the head moved, want none", n)
	}
	pr := m.pullOf(t, "p")
	if pr["question"] != "merge-p-2" || pr["asked_commit"] != moved {
		t.Fatalf("pull request = %v, want merge-p-2 asked about %s", pr, moved)
	}

	m.answerQuestion(t, "merge-p-2", "merge", "Merge it now.")
	d = m.out(t, m.check(t, mergeClock, ""))["pass:p"]
	merges := m.merges(t)
	if d.Decision != "merge" || len(merges) != 1 || !slices.Contains(merges[0], moved) {
		t.Fatalf("decision %+v, merge calls %v, want one merge pinned to %s", d, merges, moved)
	}
}

func TestMergeAnswerFixAgain(t *testing.T) {
	m := answersRig(t, "fx")
	sent := 0
	send := func() {
		sent++
		head := m.push(t, "fx", "fix"+strconv.Itoa(sent)+".txt")
		m.report(t, "fx", 1, head, "inside")
		m.answer(t, "fx", 1, head, failing)
		m.rowsDone(t)
	}
	for range 2 {
		send()
		m.out(t, m.check(t, workedClock, ""))
	}
	send()
	if d := m.out(t, m.check(t, workedClock, ""))["pass:fx"]; d.Decision != "ask" || d.Reason != "fix_limit" {
		t.Fatalf("third failure = %+v, want ask for fix_limit", d)
	}
	if got := len(m.launches(t)); got != 3 {
		t.Fatalf("launches = %d, want 3", got)
	}

	words := "Rename the flag.\nKeep $HOME and 'quotes' as they are."
	m.answerQuestion(t, "merge-fx-1", "fix again", words)
	m.out(t, m.check(t, workedClock, ""))
	if got := len(m.launches(t)); got != 4 {
		t.Fatalf("launches after fix again = %d, want 4", got)
	}
	if n := m.promptsWith(t, words); n != 1 {
		t.Errorf("launches carrying the words = %d, want 1", n)
	}
	if got := m.pullOf(t, "fx")["fix_attempts"]; got != float64(3) {
		t.Errorf("fix_attempts = %v, want 3", got)
	}
	m.out(t, m.check(t, workedClock, ""))
	if got := len(m.launches(t)); got != 4 {
		t.Errorf("a second check launched again: %d launches, want 4", got)
	}

	send()
	d := m.out(t, m.check(t, workedClock, ""))["pass:fx"]
	if d.Decision != "ask" || d.Reason != "fix_limit" {
		t.Fatalf("failure after fix again = %+v, want ask for fix_limit", d)
	}
	if q := m.pullOf(t, "fx")["question"]; q != "merge-fx-2" {
		t.Errorf("question = %v, want merge-fx-2", q)
	}
	text := ""
	for _, q := range m.release(t)["questions"].([]any) {
		if qm := q.(map[string]any); qm["id"] == "merge-fx-2" {
			text, _ = qm["text"].(string)
		}
	}
	if !strings.Contains(text, "failed 4 times") {
		t.Errorf("question text = %q, want it to count 4 failures", text)
	}
}

func TestMergeAnswerCloses(t *testing.T) {
	m := answersRig(t, "cl")
	rel := m.release(t)
	rel["passes"] = append(rel["passes"].([]any), map[string]any{
		"name": "next", "summary": "After.", "round": "r1", "planned_minutes": 30,
		"state": "not_started", "after": []string{"cl"},
	})
	body, err := json.MarshalIndent(rel, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, m.state, "releases/proj.json", string(body))

	m.askOutside(t, "cl")
	ruling := m.answerQuestion(t, "merge-cl-1", "close", "Close it.")

	closes := func() int {
		n := 0
		for _, a := range m.ghCalls(t) {
			if len(a) > 1 && a[0] == "pr" && a[1] == "close" {
				n++
			}
		}
		return n
	}

	rows := m.readRows(t)
	for _, r := range rows {
		r["state"], r["status"] = "working", "busy"
	}
	m.setRows(t, rows...)
	m.control(t, "stop-fails", true)
	if got := m.check(t, mergeClock, ""); got.code != 2 {
		t.Fatalf("check with a failing stop: exit %d, want 2; stderr %s", got.code, got.stderr)
	}
	if got := m.passRecord(t, "cl")["state"]; got != "building" {
		t.Fatalf("pass state after the failed stop = %v, want building", got)
	}
	if n := closes(); n != 1 {
		t.Fatalf("gh pr close calls = %d, want 1", n)
	}

	m.control(t, "stop-fails", false)
	out := m.out(t, m.check(t, mergeClock, ""))
	if d := out["pass:cl"]; d.Decision != "close" {
		t.Fatalf("decision = %+v, want close", d)
	}
	if n := closes(); n != 1 {
		t.Errorf("gh pr close calls after the second run = %d, want it skipped", n)
	}
	pass := m.passRecord(t, "cl")
	closed, _ := pass["closed"].(map[string]any)
	if pass["state"] != "done" || closed["ruling"] != ruling {
		t.Fatalf("pass = %v, want done and closed by %s", pass, ruling)
	}
	if ended := m.sessionOf(t, "cl")["ended"]; ended == nil {
		t.Error("the session is not recorded as ended")
	}
	if s := readEst(t, m.state, mergeClock); len(s.WaitsForGeoff) != 0 {
		t.Errorf("waits_for_geoff = %+v, want none", s.WaitsForGeoff)
	}
	for rec, d := range out {
		if d.Decision == "start" {
			t.Errorf("%s was started after the close", rec)
		}
	}
	s := readRulingSnap(t, m.state, mergeClock)
	if free, whyNot, _, _, _ := s.pass(t, "proj", "next"); free == nil || *free || !strings.Contains(strings.Join(whyNot, " "), "cl") {
		t.Errorf("next: free %v, why_not %v, want blocked naming cl", free, whyNot)
	}
	if got := m.passRecord(t, "next")["state"]; got != "not_started" {
		t.Errorf("next state = %v, want not_started", got)
	}
}
