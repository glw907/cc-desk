package e2e

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

const answeringRuling = `{"format":1,"id":"r-0002","time":"2026-11-20T22:30:00Z","desk":"earlier","words":"Yes.","project":"north","round":"n1","question":"north-q1","option":"yes"}`

func replaceArgs(replaces string) []string {
	return []string{"ruling", "record", "--project", "north", "--round", "n1", "--question", "north-q1", "--option", "no", "--replaces", replaces, "--", rulingWords}
}

// answerQ1 writes r-0002 as north-q1's answer, before the verb runs.
func answerQ1(t *testing.T, dir string) func() {
	return func() {
		writeIn(t, dir, "rulings.jsonl", earlierRuling+"\n"+answeringRuling+"\n")
		path := filepath.Join(dir, "releases/north.json")
		old := `"recommended": "yes"},` + "\n" + `    {"id": "north-q2"`
		body := readFile(t, path)
		if !strings.Contains(body, old) {
			t.Fatal("the order fixture's north-q1 line moved")
		}
		body = strings.Replace(body, old, `"recommended": "yes", "answer": {"ruling": "r-0002"}},`+"\n"+`    {"id": "north-q2"`, 1)
		writeIn(t, dir, "releases/north.json", body)
	}
}

func TestReplace(t *testing.T) {
	const deskSession = "e2e-desk-session"

	t.Run("replaces_before_the_pass_starts", func(t *testing.T) {
		dir := newStateFrom(t, "order")
		r := recordRulingArgs(t, inTerminalDesk, deskSession, dir, at(t, 20), replaceArgs("r-0002"), answerQ1(t, dir))
		if r.code != 0 || !strings.Contains(r.stderr, "recorded ruling r-0003") {
			t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
		}
		lines := strings.Split(strings.TrimSuffix(readFile(t, filepath.Join(dir, "rulings.jsonl")), "\n"), "\n")
		if len(lines) != 3 || lines[1] != answeringRuling {
			t.Fatalf("rulings.jsonl = %q, want the earlier lines kept", lines)
		}
		var got state.Ruling
		if err := json.Unmarshal([]byte(lines[2]), &got); err != nil {
			t.Fatal(err)
		}
		if got.ID != "r-0003" || got.Replaces != "r-0002" || got.Option != "no" {
			t.Errorf("ruling = %+v", got)
		}
		if !strings.Contains(readFile(t, filepath.Join(dir, "releases/north.json")), `"ruling": "r-0003"`) {
			t.Error("north-q1's answer does not name r-0003")
		}
		s := readRulingSnap(t, dir, at(t, 21))
		free, whyNot, _, answered, kind := s.pass(t, "north", "n-big")
		if free == nil || !*free || len(whyNot) != 0 || !answered || kind == nil || *kind != "ruling" {
			t.Errorf("n-big: free %v, why_not %v, answered %v, kind %v", free, whyNot, answered, kind)
		}
	})

	t.Run("refuses_a_missing_ruling", func(t *testing.T) {
		dir := newStateFrom(t, "order")
		r := recordRulingArgs(t, inTerminalDesk, deskSession, dir, at(t, 20), replaceArgs("r-0009"), answerQ1(t, dir))
		if r.code != 2 || !strings.Contains(r.stderr, `ruling "r-0009" does not exist`) {
			t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
		}
		if got := readFile(t, filepath.Join(dir, "rulings.jsonl")); strings.Count(got, "\n") != 2 {
			t.Errorf("rulings.jsonl = %q, want no new line", got)
		}
	})

	t.Run("refuses_a_started_pass", func(t *testing.T) {
		dir := newStateFrom(t, "order")
		answer := answerQ1(t, dir)
		r := recordRulingArgs(t, inTerminalDesk, deskSession, dir, at(t, 20), replaceArgs("r-0002"), func() {
			answer()
			path := filepath.Join(dir, "releases/north.json")
			body := strings.Replace(readFile(t, path), `"planned_minutes": 400, "state": "not_started"`, `"planned_minutes": 400, "state": "building", "started": "2026-11-20T21:00:00Z"`, 1)
			writeIn(t, dir, "releases/north.json", body)
		})
		if r.code != 2 || !strings.Contains(r.stderr, "has started; replacing it while it builds comes in pass 3") {
			t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
		}
	})
}
