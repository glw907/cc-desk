//go:build unix

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// startLogLines gives the decision and outcome of each check-log line for the
// record.
func startLogLines(t *testing.T, dir, record string) [][2]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "check-log.jsonl"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out [][2]string
	for line := range strings.SplitSeq(strings.TrimSpace(string(b)), "\n") {
		var l struct {
			Record   string `json:"record"`
			Decision string `json:"decision"`
			Outcome  string `json:"outcome"`
		}
		if err := json.Unmarshal([]byte(line), &l); err != nil {
			t.Fatalf("check-log.jsonl line %q: %v", line, err)
		}
		if l.Record == record {
			out = append(out, [2]string{l.Decision, l.Outcome})
		}
	}
	return out
}

// A pass held by an unanswered question waits; once a ruling answers it the
// check starts the pass, logs the start, and launches it in the same run.
func TestCheckStartsAFreePass(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")

	var rel map[string]any
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "releases/proj.json")), &rel); err != nil {
		t.Fatal(err)
	}
	rel["passes"].([]any)[0].(map[string]any)["questions"] = []string{"q1"}
	rel["questions"] = []any{map[string]any{"id": "q1", "text": "Go?", "options": []string{"yes", "no"}, "recommended": "yes"}}
	body, err := json.Marshal(rel)
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, r.state, "releases/proj.json", string(body))

	r.check(t, minutes(t, 1))
	if got := r.passRecord(t, "a")["state"]; got != "not_started" {
		t.Fatalf("with q1 open the pass is %v, want not_started", got)
	}
	if lines := startLogLines(t, r.state, "pass:a"); len(lines) != 0 {
		t.Fatalf("log lines %v for a held pass, want none", lines)
	}

	writeIn(t, r.state, "rulings.jsonl", `{"format":1,"id":"r-0001","time":"2026-11-20T22:00:00Z","desk":"d","words":"Go.","project":"proj","round":"r1","question":"q1","option":"yes"}`+"\n")
	rel["questions"].([]any)[0].(map[string]any)["answer"] = map[string]any{"ruling": "r-0001"}
	if body, err = json.Marshal(rel); err != nil {
		t.Fatal(err)
	}
	writeIn(t, r.state, "releases/proj.json", string(body))

	got := r.check(t, minutes(t, 2))
	if st := r.passRecord(t, "a")["state"]; st != "building" {
		t.Fatalf("after the ruling the pass is %v, want building", st)
	}
	if !strings.Contains(got.stdout, `"decision":"start"`) {
		t.Errorf("check output %q lacks start", got.stdout)
	}
	lines := startLogLines(t, r.state, "pass:a")
	if len(lines) != 2 || lines[0] != [2]string{"start", "started"} || lines[1][0] != "launch" || !strings.HasPrefix(lines[1][1], "launched ") {
		t.Errorf("log lines %v, want start/started then launch/launched", lines)
	}
	if n := len(r.launches(t)); n != 1 {
		t.Errorf("%d launches, want 1", n)
	}
}
