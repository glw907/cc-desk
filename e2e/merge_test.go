//go:build unix

package e2e

import (
	"strings"
	"testing"
)

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
