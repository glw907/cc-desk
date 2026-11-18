package e2e

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// refusalNames are the four rules a refused start can name, in the order
// the design lists them.
var refusalNames = []string{
	"unfinished work before it",
	"an unanswered question",
	"the one-pass rule",
	"the 300-minute rule",
}

// at returns the worked clock moved on by minutes, so each step of a test
// runs later than the one before.
func at(t *testing.T, minutes int) string {
	t.Helper()
	base, err := time.Parse(time.RFC3339, workedClock)
	if err != nil {
		t.Fatal(err)
	}
	return base.Add(time.Duration(minutes) * time.Minute).Format(time.RFC3339)
}

// namesIn returns which of the four rule names the text holds.
func namesIn(text string) []string {
	var got []string
	for _, n := range refusalNames {
		if strings.Contains(text, n) {
			got = append(got, n)
		}
	}
	return got
}

// requireRefusal fails unless r is a refusal (exit 2) that names exactly
// the given rules.
func requireRefusal(t *testing.T, r result, want ...string) {
	t.Helper()
	if r.code != 2 {
		t.Fatalf("exit %d, want 2; stderr: %s", r.code, r.stderr)
	}
	if got := namesIn(r.stderr); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("refusal names %q, want exactly %q; stderr: %s", got, want, r.stderr)
	}
}

// orderSnap is the part of the snapshot the order tests read.
type orderSnap struct {
	Releases []struct {
		Project string `json:"project"`
		Rounds  []struct {
			Passes []struct {
				Name  string `json:"name"`
				State string `json:"state"`
			} `json:"passes"`
		} `json:"rounds"`
	} `json:"releases"`
	NextSession struct {
		State string `json:"state"`
	} `json:"next_session"`
}

func (s orderSnap) stateOf(t *testing.T, project, pass string) string {
	t.Helper()
	for _, rel := range s.Releases {
		if rel.Project != project {
			continue
		}
		for _, rd := range rel.Rounds {
			for _, p := range rd.Passes {
				if p.Name == pass {
					return p.State
				}
			}
		}
	}
	t.Fatalf("pass %s/%s not in the snapshot", project, pass)
	return ""
}

func (s orderSnap) building() []string {
	var out []string
	for _, rel := range s.Releases {
		for _, rd := range rel.Rounds {
			for _, p := range rd.Passes {
				if p.State == "building" {
					out = append(out, rel.Project+"/"+p.Name)
				}
			}
		}
	}
	return out
}

func readOrderSnap(t *testing.T, dir, now string) orderSnap {
	t.Helper()
	r := run(t, dir, now, "snapshot")
	if r.code != 0 {
		t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
	}
	var s orderSnap
	if err := json.Unmarshal([]byte(r.stdout), &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestWaitingOrder(t *testing.T) {
	// The cases on copy A run in order on one folder.
	a := newStateFrom(t, "order")

	t.Run("free_and_question", func(t *testing.T) {
		if r := run(t, a, at(t, 0), "pass", "start", "north", "n-free"); r.code != 0 {
			t.Fatalf("exit %d: %s", r.code, r.stderr)
		}
		if got := readOrderSnap(t, a, at(t, 1)).stateOf(t, "north", "n-free"); got != "building" {
			t.Fatalf("n-free is %s, want building", got)
		}
		before := folder(t, a)
		requireRefusal(t, run(t, a, at(t, 2), "pass", "start", "north", "n-guess"), "an unanswered question")
		sameFolder(t, before, folder(t, a))
	})

	t.Run("too_big", func(t *testing.T) {
		b := newStateFrom(t, "order")
		before := folder(t, b)
		requireRefusal(t, run(t, b, at(t, 0), "pass", "start", "north", "n-big", "--guess", "north-q1"), "the 300-minute rule")
		sameFolder(t, before, folder(t, b))
		s := readOrderSnap(t, b, at(t, 1))
		if got := s.building(); len(got) != 0 {
			t.Errorf("building %q, want none", got)
		}
		if s.NextSession.State != "ready_now" {
			t.Errorf("next_session.state = %q, want ready_now", s.NextSession.State)
		}
	})

	t.Run("guess_starts", func(t *testing.T) {
		if r := run(t, a, at(t, 10), "pass", "start", "north", "n-guess", "--guess", "north-q2"); r.code != 0 {
			t.Fatalf("exit %d: %s", r.code, r.stderr)
		}
		if got := readOrderSnap(t, a, at(t, 11)).stateOf(t, "north", "n-guess"); got != "building" {
			t.Fatalf("n-guess is %s, want building", got)
		}
		var rel struct {
			Passes []struct {
				Name  string `json:"name"`
				Guess *struct {
					Question       string `json:"question"`
					Recommendation string `json:"recommendation"`
					At             string `json:"at"`
				} `json:"guess"`
			} `json:"passes"`
		}
		if err := json.Unmarshal([]byte(folder(t, a)["releases/north.json"]), &rel); err != nil {
			t.Fatal(err)
		}
		for _, p := range rel.Passes {
			if p.Name != "n-guess" {
				continue
			}
			g := p.Guess
			if g == nil || g.Question != "north-q2" || g.Recommendation != "yes" || g.At != at(t, 10) {
				t.Errorf("guess = %+v, want north-q2, yes, at %s", g, at(t, 10))
			}
			return
		}
		t.Error("north.json has no n-guess")
	})

	t.Run("other_project", func(t *testing.T) {
		before := folder(t, a)
		requireRefusal(t, run(t, a, at(t, 20), "pass", "start", "south", "s-guess", "--guess", "south-q1"), "the 300-minute rule")
		sameFolder(t, before, folder(t, a))
	})

	t.Run("one_pass", func(t *testing.T) {
		if r := run(t, a, at(t, 30), "pass", "finish", "north", "n-guess"); r.code != 0 {
			t.Fatalf("finish exit %d: %s", r.code, r.stderr)
		}
		before := folder(t, a)
		requireRefusal(t, run(t, a, at(t, 31), "pass", "start", "north", "n-after"), "the one-pass rule")
		sameFolder(t, before, folder(t, a))
	})
}
