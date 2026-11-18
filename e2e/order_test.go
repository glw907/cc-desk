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
		var open struct {
			OpenGuesses []struct {
				Project  string `json:"project"`
				Pass     string `json:"pass"`
				Question string `json:"question"`
				Minutes  int    `json:"minutes"`
			} `json:"open_guesses"`
			Total int `json:"guess_minutes_total"`
		}
		if err := json.Unmarshal([]byte(run(t, a, at(t, 11), "snapshot").stdout), &open); err != nil {
			t.Fatal(err)
		}
		if len(open.OpenGuesses) != 1 || open.OpenGuesses[0].Pass != "n-guess" || open.OpenGuesses[0].Question != "north-q2" || open.Total != 200 {
			t.Errorf("open_guesses %+v total %d, want n-guess on north-q2 and 200", open.OpenGuesses, open.Total)
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

// limitSnap is the part of the snapshot the limit tests read.
type limitSnap struct {
	Waiting []struct {
		Project   string   `json:"project"`
		Pass      string   `json:"pass"`
		Limits    []string `json:"limits"`
		ResetText *string  `json:"reset_text"`
	} `json:"waiting"`
	NextSession struct {
		State   string   `json:"state"`
		Reasons []string `json:"reasons"`
	} `json:"next_session"`
	Releases []struct {
		Project string `json:"project"`
		Rounds  []struct {
			Passes []struct {
				Name  string `json:"name"`
				State string `json:"state"`
			} `json:"passes"`
		} `json:"rounds"`
	} `json:"releases"`
}

func (s limitSnap) stateOf(t *testing.T, project, pass string) string {
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

// waitingOn returns the limits a pass is listed as waiting on, or nil.
func (s limitSnap) waitingOn(project, pass string) []string {
	for _, w := range s.Waiting {
		if w.Project == project && w.Pass == pass {
			return w.Limits
		}
	}
	return nil
}

func snapWith(t *testing.T, si *standIn, dir, now string) limitSnap {
	t.Helper()
	r := si.run(t, dir, now, "snapshot")
	if r.code != 0 {
		t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
	}
	var s limitSnap
	if err := json.Unmarshal([]byte(r.stdout), &s); err != nil {
		t.Fatal(err)
	}
	return s
}

// start runs a start that must exit 0 and returns its stderr.
func start(t *testing.T, si *standIn, dir, now string, args ...string) string {
	t.Helper()
	r := si.run(t, dir, now, append([]string{"pass", "start"}, args...)...)
	if r.code != 0 {
		t.Fatalf("start %v exit %d: %s", args, r.code, r.stderr)
	}
	return r.stderr
}

func requireState(t *testing.T, s limitSnap, project, pass, want string) {
	t.Helper()
	if got := s.stateOf(t, project, pass); got != want {
		t.Errorf("%s/%s is %s, want %s", project, pass, got, want)
	}
}

func requireWaitingOn(t *testing.T, s limitSnap, project, pass string, want ...string) {
	t.Helper()
	requireState(t, s, project, pass, "waiting")
	if got := s.waitingOn(project, pass); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s/%s waits on %v, want %v", project, pass, got, want)
	}
}

const resetText = "Nov 20, 10:59pm (America/Anchorage)"

func TestLimits(t *testing.T) {
	t.Run("two_pass_limit", func(t *testing.T) {
		si := newStandIn(t, 0, usageResult(40))
		dir := newStateFrom(t, "order")
		start(t, si, dir, at(t, 0), "north", "n-free")
		start(t, si, dir, at(t, 1), "south", "s-free")
		if got := start(t, si, dir, at(t, 2), "north", "n-two"); !strings.Contains(got, "waiting north/n-two on two_pass_limit") {
			t.Errorf("wait line = %q", got)
		}
		start(t, si, dir, at(t, 3), "south", "s-two")
		s := snapWith(t, si, dir, at(t, 4))
		requireWaitingOn(t, s, "north", "n-two", "two_pass_limit")
		requireWaitingOn(t, s, "south", "s-two", "two_pass_limit")
		if len(s.Waiting) != 2 || s.Waiting[0].Pass != "n-two" {
			t.Errorf("waiting = %+v, want n-two first", s.Waiting)
		}

		if r := si.run(t, dir, at(t, 5), "pass", "finish", "north", "n-free"); r.code != 0 {
			t.Fatalf("finish exit %d: %s", r.code, r.stderr)
		}
		s = snapWith(t, si, dir, at(t, 6))
		requireState(t, s, "north", "n-two", "building")
		requireWaitingOn(t, s, "south", "s-two", "two_pass_limit")
	})

	t.Run("usage_99", func(t *testing.T) {
		si := newStandIn(t, 0, usageResult(99))
		dir := newStateFrom(t, "order")
		start(t, si, dir, at(t, 0), "north", "n-free")
		requireState(t, snapWith(t, si, dir, at(t, 1)), "north", "n-free", "building")
	})

	t.Run("usage_100_then_99", func(t *testing.T) {
		si := newStandIn(t, 0, usageResult(100))
		dir := newStateFrom(t, "order")
		start(t, si, dir, at(t, 0), "north", "n-free")
		s := snapWith(t, si, dir, at(t, 1))
		requireWaitingOn(t, s, "north", "n-free", "usage")
		if len(s.Waiting) != 1 || s.Waiting[0].ResetText == nil || *s.Waiting[0].ResetText != resetText {
			t.Errorf("waiting = %+v, want reset text %q", s.Waiting, resetText)
		}

		si.set(t, usageResult(99))
		requireState(t, snapWith(t, si, dir, at(t, 6)), "north", "n-free", "building")
	})

	t.Run("stop_85", func(t *testing.T) {
		si := newStandIn(t, 0, usageResult(87))
		dir := newStateFrom(t, "order")
		writeIn(t, dir, "settings.json", `{"format": 1, "usage_stop_percent": 85}`+"\n")
		start(t, si, dir, at(t, 0), "north", "n-free")
		requireWaitingOn(t, snapWith(t, si, dir, at(t, 1)), "north", "n-free", "usage")
	})

	t.Run("unparseable", func(t *testing.T) {
		si := newStandIn(t, 0, "this is not a result\n")
		dir := newStateFrom(t, "order")
		start(t, si, dir, at(t, 0), "north", "n-free")
		s := snapWith(t, si, dir, at(t, 1))
		requireWaitingOn(t, s, "north", "n-free", "usage")
		if len(s.Waiting) != 1 || s.Waiting[0].ResetText != nil {
			t.Errorf("waiting = %+v, want a null reset text", s.Waiting)
		}
		if s.NextSession.State != "needs_you_now" || strings.Join(s.NextSession.Reasons, "|") != "the usage figure can't be read" {
			t.Errorf("next_session = %+v", s.NextSession)
		}
	})

	t.Run("cache", func(t *testing.T) {
		si := newStandIn(t, 0, usageResult(40))
		dir := newStateFrom(t, "order")
		snapWith(t, si, dir, at(t, 0))
		snapWith(t, si, dir, at(t, 1))
		if got := si.calls(t); got != 1 {
			t.Errorf("after two verbs 1 minute apart the stand-in ran %d times, want 1", got)
		}
		snapWith(t, si, dir, at(t, 5))
		if got := si.calls(t); got != 2 {
			t.Errorf("after a verb 5 minutes on the stand-in ran %d times, want 2", got)
		}
	})
}

func TestGuessWaits(t *testing.T) {
	si := newStandIn(t, 0, usageResult(40))
	dir := newStateFrom(t, "order")
	start(t, si, dir, at(t, 0), "north", "n-free")
	start(t, si, dir, at(t, 1), "south", "s-free")
	got := start(t, si, dir, at(t, 2), "north", "n-guess", "--guess", "north-q2")
	if !strings.Contains(got, "waiting north/n-guess on two_pass_limit") {
		t.Errorf("wait line = %q", got)
	}

	var rel struct {
		Passes []struct {
			Name  string `json:"name"`
			Guess *struct {
				Question string `json:"question"`
			} `json:"guess"`
		} `json:"passes"`
	}
	if err := json.Unmarshal([]byte(folder(t, dir)["releases/north.json"]), &rel); err != nil {
		t.Fatal(err)
	}
	var recorded bool
	for _, p := range rel.Passes {
		if p.Name == "n-guess" && p.Guess != nil && p.Guess.Question == "north-q2" {
			recorded = true
		}
	}
	if !recorded {
		t.Error("the waiting guess start did not record its guess")
	}

	r := si.run(t, dir, at(t, 3), "snapshot")
	if r.code != 0 {
		t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
	}
	var open struct {
		OpenGuesses []struct {
			Pass string `json:"pass"`
		} `json:"open_guesses"`
		Total int `json:"guess_minutes_total"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &open); err != nil {
		t.Fatal(err)
	}
	if len(open.OpenGuesses) != 1 || open.OpenGuesses[0].Pass != "n-guess" || open.Total != 200 {
		t.Errorf("open_guesses %+v total %d, want n-guess and 200", open.OpenGuesses, open.Total)
	}
}

func TestFreeVerb(t *testing.T) {
	si := newStandIn(t, 0, usageResult(40))
	dir := newStateFrom(t, "order")
	start(t, si, dir, at(t, 0), "north", "n-free")
	start(t, si, dir, at(t, 1), "south", "s-free")
	start(t, si, dir, at(t, 2), "north", "n-two")
	r := si.run(t, dir, at(t, 3), "free")
	if r.code != 0 {
		t.Fatalf("free exit %d: %s", r.code, r.stderr)
	}
	lines := strings.Split(strings.TrimSpace(r.stdout), "\n")
	for _, want := range []string{
		"free south/s-two",
		"waiting north/n-two on two_pass_limit",
		"not free north/n-big: an unanswered question",
	} {
		var found bool
		for _, l := range lines {
			if l == want {
				found = true
			}
		}
		if !found {
			t.Errorf("free printed no line %q; lines: %q", want, lines)
		}
	}
}

func TestUsageReader(t *testing.T) {
	built := `{"type":"result","subtype":"success","is_error":false,"num_turns":0,"total_cost_usd":0,` +
		`"result":"Current week (Fable): 12% used · resets Nov 21, 1:00am (America/Anchorage)\n` +
		`Current session: 96% used · resets 3:00pm\n` +
		`Current week (all models): 61% used · resets Nov 20, 10:59pm (America/Anchorage)\n"}` + "\n"
	si := newStandIn(t, 0, built)
	dir := newStateFrom(t, "order")
	r := si.run(t, dir, at(t, 0), "usage")
	if r.code != 0 {
		t.Fatalf("usage exit %d: %s", r.code, r.stderr)
	}
	if want := "61% used, resets " + resetText + "\n"; r.stdout != want {
		t.Errorf("usage printed %q, want %q", r.stdout, want)
	}
}
