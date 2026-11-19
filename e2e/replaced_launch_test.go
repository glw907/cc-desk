//go:build unix

package e2e

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

const (
	replacedQuestion = "Should the flag be renamed?"
	replacedWords    = "Keep the old flag name.\nAdd an alias."
)

// markReplaced gives the pass a question, answers it with a ruling that
// carries replacedWords, and marks the pass as having its answer replaced.
func (m mergeRig) markReplaced(t *testing.T, pass string) {
	t.Helper()
	rel := m.release(t)
	rel["questions"] = []any{map[string]any{
		"id": "proj-q1", "text": replacedQuestion, "options": []string{"yes", "no"}, "recommended": "yes",
	}}
	for _, p := range rel["passes"].([]any) {
		if pm := p.(map[string]any); pm["name"] == pass {
			pm["questions"] = []string{"proj-q1"}
		}
	}
	body, err := json.MarshalIndent(rel, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, m.state, "releases/proj.json", string(body))
	id := m.answerQuestion(t, "proj-q1", "no", replacedWords)

	rel = m.release(t)
	for _, p := range rel["passes"].([]any) {
		if pm := p.(map[string]any); pm["name"] == pass {
			pm["replaced"] = map[string]any{"ruling": id, "time": "2026-11-20T23:45:00Z"}
		}
	}
	if body, err = json.MarshalIndent(rel, "", "  "); err != nil {
		t.Fatal(err)
	}
	writeIn(t, m.state, "releases/proj.json", string(body))
}

// countedRelaunches counts the relaunches of the pass's session that count
// toward the two, and gives the total as well.
func (m mergeRig) countedRelaunches(t *testing.T, pass string) (counted, total int) {
	t.Helper()
	rel, _ := m.sessionOf(t, pass)["relaunches"].([]any)
	for _, r := range rel {
		if r.(map[string]any)["counted"] == true {
			counted++
		}
	}
	return counted, len(rel)
}

func TestReplacedPassIsCarriedOut(t *testing.T) {
	m := newMergeRig(t, "rp")
	old := m.sessionOf(t, "rp")["short_id"].(string)
	head := m.push(t, "rp", "rp.txt")
	m.report(t, "rp", 1, head, "inside")
	m.answer(t, "rp", 1, head, passing)
	m.markReplaced(t, "rp")
	counted, total := m.countedRelaunches(t, "rp")
	before := len(m.calls(t))

	d := m.out(t, m.check(t, mergeClock, ""))["pass:rp"]
	if d.Decision != "end_and_relaunch" || d.Reason != "replaced" {
		t.Fatalf("decision = %+v, want end_and_relaunch for replaced", d)
	}

	got := m.calls(t)[before:]
	var kinds []string
	for _, c := range got {
		kinds = append(kinds, c.Args[0])
	}
	stop := slices.Index(kinds, "stop")
	if stop < 0 || got[stop].Args[1] != old {
		t.Fatalf("calls %v: want claude stop %s", kinds, old)
	}
	if stop+2 >= len(kinds) || kinds[stop+1] != "agents" || !slices.Contains(kinds[stop+2:], "--bg") {
		t.Errorf("calls %v: want stop, then a list read, then the launch", kinds)
	}

	launches := m.launches(t)
	if len(launches) != 2 {
		t.Fatalf("%d launches, want 2", len(launches))
	}
	prompt := relaunchPrompt(launches[1])
	for _, want := range []string{"Geoff replaced his answer to this question: " + replacedQuestion, replacedWords} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the relaunch prompt lacks %q:\n%s", want, prompt)
		}
	}

	c, n := m.countedRelaunches(t, "rp")
	if n != total+1 || c != counted {
		t.Errorf("relaunches %d counted %d, want %d and %d", n, c, total+1, counted)
	}
	rel, _ := m.sessionOf(t, "rp")["relaunches"].([]any)
	if last := rel[len(rel)-1].(map[string]any); last["counted"] == true || last["cause"] != "replaced" {
		t.Errorf("last relaunch = %v, want uncounted with cause replaced", last)
	}

	pass := m.passRecord(t, "rp")
	if _, ok := pass["replaced"]; ok {
		t.Errorf("the replaced mark stands after the launch: %v", pass["replaced"])
	}
	if _, ok := m.pullOf(t, "rp")["ready"]; ok {
		t.Error("the ready report stands after the launch")
	}
	if merges := m.merges(t); len(merges) != 0 {
		t.Errorf("a ready report under the mark merged: %v", merges)
	}
}

func TestReplacedPassUsageRelaunchCarriesTheParagraph(t *testing.T) {
	m := newMergeRig(t, "rp")
	requireRun(t, m.do(t, minutes(t, 3), "pass", "stop", "proj", "rp", "--reason", "usage"))
	m.markReplaced(t, "rp")

	m.out(t, m.check(t, minutes(t, 4), ""))
	launches := m.launches(t)
	if len(launches) != 2 {
		t.Fatalf("%d launches after the usage stop, want 2", len(launches))
	}
	prompt := relaunchPrompt(launches[1])
	for _, want := range []string{"Geoff replaced his answer to this question: " + replacedQuestion, replacedWords} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the usage relaunch prompt lacks %q:\n%s", want, prompt)
		}
	}
	if _, ok := m.passRecord(t, "rp")["replaced"]; ok {
		t.Error("the replaced mark stands after the usage relaunch")
	}
}
