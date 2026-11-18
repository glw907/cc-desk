package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/state"
)

// rulingWords has quotes and a non-ASCII letter, to show the words reach the
// log byte for byte through a shell's expansion.
const rulingWords = `Use the "short form" for the café list.`

const earlierRuling = `{"format":1,"id":"r-0001","time":"2026-11-20T22:00:00Z","desk":"earlier","words":"Before."}`

// rulingSnap is the part of the snapshot the ruling tests read.
type rulingSnap struct {
	Releases []struct {
		Project string `json:"project"`
		Rounds  []struct {
			Passes []struct {
				Name         string   `json:"name"`
				Free         *bool    `json:"free"`
				WhyNot       []string `json:"why_not"`
				GuessOutcome *string  `json:"guess_outcome"`
				Questions    []struct {
					ID         string  `json:"id"`
					Answered   bool    `json:"answered"`
					AnswerKind *string `json:"answer_kind"`
				} `json:"questions"`
			} `json:"passes"`
		} `json:"rounds"`
	} `json:"releases"`
	Rulings []struct {
		ID       string `json:"id"`
		Question string `json:"question"`
		Option   string `json:"option"`
	} `json:"rulings"`
	OpenGuesses       []json.RawMessage `json:"open_guesses"`
	GuessMinutesTotal int               `json:"guess_minutes_total"`
}

func (s rulingSnap) pass(t *testing.T, project, name string) (free *bool, whyNot []string, outcome *string, answered bool, kind *string) {
	t.Helper()
	for _, rel := range s.Releases {
		if rel.Project != project {
			continue
		}
		for _, rd := range rel.Rounds {
			for _, p := range rd.Passes {
				if p.Name != name {
					continue
				}
				if len(p.Questions) > 0 {
					answered, kind = p.Questions[0].Answered, p.Questions[0].AnswerKind
				}
				return p.Free, p.WhyNot, p.GuessOutcome, answered, kind
			}
		}
	}
	t.Fatalf("pass %s/%s not in the snapshot", project, name)
	return
}

func readRulingSnap(t *testing.T, dir, now string) rulingSnap {
	t.Helper()
	r := run(t, dir, now, "snapshot")
	if r.code != 0 {
		t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
	}
	var s rulingSnap
	if err := json.Unmarshal([]byte(r.stdout), &s); err != nil {
		t.Fatal(err)
	}
	return s
}

// rulingArgs is the verb's argument list after the global flags.
func rulingArgs(question, option string) []string {
	return []string{"ruling", "record", "--project", "north", "--round", "n1", "--question", question, "--option", option, "--", rulingWords}
}

func globalArgs(t *testing.T, dir, now string) []string {
	t.Helper()
	return []string{"--state-dir", dir, "--now", now, "--usage-command", usageStandIn(t)}
}

// how the ruling verb is run.
const (
	inTerminalDesk = iota // under a desk that has a terminal, as a desk's command does
	inSetsidDesk          // under a desk that has none
	outsideDesk           // by the test directly
)

// recordRuling registers a desk, runs the ruling verb as how says with
// sessionID in CLAUDE_CODE_SESSION_ID (last in the environment), and returns
// the verb's result. The desk's own session id is always registered. A
// non-nil beforeRun is called once desks.json and rulings.jsonl are in place
// and just before the verb runs.
func recordRuling(t *testing.T, how int, sessionID, dir, now, question, option string, beforeRun func()) result {
	t.Helper()
	return recordRulingArgs(t, how, sessionID, dir, now, rulingArgs(question, option), beforeRun)
}

// recordRulingArgs is recordRuling with the verb's own argument list.
func recordRulingArgs(t *testing.T, how int, sessionID, dir, now string, verb []string, beforeRun func()) result {
	t.Helper()
	const deskSession = "e2e-desk-session"
	out := t.TempDir()
	args := append(globalArgs(t, dir, now), verb...)

	script := ""
	if how != outsideDesk {
		// The verb runs in the ( ... ) & form under setsid, so it has no
		// terminal of its own and the desk shell is its ancestor.
		script = `( setsid -w "$BIN" "$@" > "$OUT/stdout" 2> "$OUT/stderr"; echo $? > "$OUT/code" ) &`
		script = `set -- ` + shellWords(args) + "\n" + script
	}
	env := []string{"BIN=" + binary, "OUT=" + out, "CLAUDE_CODE_SESSION_ID=" + sessionID}
	pid, release, wait := startDesk(t, how != inSetsidDesk, env, script)

	entry, err := desk.Entry(deskSession, pid)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state.Desks{Format: state.Format, Desks: []state.Desk{entry}})
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, dir, "desks.json", string(data)+"\n")
	writeIn(t, dir, "rulings.jsonl", earlierRuling)
	if beforeRun != nil {
		beforeRun()
	}

	if how == outsideDesk {
		cmd := exec.Command(binary, args...)
		cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin", "CLAUDE_CODE_SESSION_ID="+sessionID)
		var so, se bytes.Buffer
		cmd.Stdout, cmd.Stderr = &so, &se
		code := 0
		if err := cmd.Run(); err != nil {
			ee, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			code = ee.ExitCode()
		}
		return result{so.String(), se.String(), code}
	}
	release()
	wait()
	code, err := strconv.Atoi(strings.TrimSpace(readFile(t, filepath.Join(out, "code"))))
	if err != nil {
		t.Fatalf("the desk's verb left no exit code: %v", err)
	}
	return result{readFile(t, filepath.Join(out, "stdout")), readFile(t, filepath.Join(out, "stderr")), code}
}

// shellWords single-quotes each word for an sh command line.
func shellWords(words []string) string {
	q := make([]string, len(words))
	for i, w := range words {
		q[i] = "'" + strings.ReplaceAll(w, "'", `'\''`) + "'"
	}
	return strings.Join(q, " ")
}

func TestRuling(t *testing.T) {
	const deskSession = "e2e-desk-session"

	t.Run("answers_and_frees", func(t *testing.T) {
		dir := newStateFrom(t, "order")
		before := readRulingSnap(t, dir, at(t, 0))
		if free, _, _, _, _ := before.pass(t, "north", "n-big"); free == nil || *free {
			t.Fatal("n-big is free before its question is answered")
		}
		r := recordRuling(t, inTerminalDesk, deskSession, dir, at(t, 20), "north-q1", "yes", nil)
		if r.code != 0 || !strings.Contains(r.stderr, "recorded ruling r-0002") {
			t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
		}
		lines := strings.Split(readFile(t, filepath.Join(dir, "rulings.jsonl")), "\n")
		if len(lines) != 3 || lines[0] != earlierRuling || lines[2] != "" {
			t.Fatalf("rulings.jsonl = %q", lines)
		}
		var got state.Ruling
		if err := json.Unmarshal([]byte(lines[1]), &got); err != nil {
			t.Fatal(err)
		}
		if got.ID != "r-0002" || got.Words != rulingWords || got.Desk != deskSession {
			t.Errorf("ruling = %+v", got)
		}
		if !strings.Contains(readFile(t, filepath.Join(dir, "releases/north.json")), `"ruling": "r-0002"`) {
			t.Error("north-q1's answer does not name r-0002")
		}
		s := readRulingSnap(t, dir, at(t, 21))
		free, whyNot, _, answered, kind := s.pass(t, "north", "n-big")
		if free == nil || !*free || len(whyNot) != 0 || !answered || kind == nil || *kind != "ruling" {
			t.Errorf("n-big: free %v, why_not %v, answered %v, kind %v", free, whyNot, answered, kind)
		}
		found := false
		for _, r := range s.Rulings {
			if r.ID == "r-0002" && r.Question == "north-q1" && r.Option == "yes" {
				found = true
			}
		}
		if !found {
			t.Errorf("snapshot rulings = %+v", s.Rulings)
		}
	})

	for _, c := range []struct{ name, option, outcome string }{
		{"guess_held", "yes", "held"},
		{"guess_missed", "no", "missed"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := newStateFrom(t, "order")
			if r := run(t, dir, at(t, 10), "pass", "start", "north", "n-guess", "--guess", "north-q2"); r.code != 0 {
				t.Fatalf("start on a guess: exit %d: %s", r.code, r.stderr)
			}
			if s := readRulingSnap(t, dir, at(t, 11)); s.GuessMinutesTotal != 200 || len(s.OpenGuesses) != 1 {
				t.Fatalf("before: total %d, %d open guesses; want 200 and 1", s.GuessMinutesTotal, len(s.OpenGuesses))
			}
			r := recordRuling(t, inTerminalDesk, deskSession, dir, at(t, 20), "north-q2", c.option, nil)
			if r.code != 0 {
				t.Fatalf("exit %d: %s", r.code, r.stderr)
			}
			s := readRulingSnap(t, dir, at(t, 21))
			if _, _, outcome, _, _ := s.pass(t, "north", "n-guess"); outcome == nil || *outcome != c.outcome {
				t.Errorf("guess_outcome = %v, want %q", outcome, c.outcome)
			}
			if s.GuessMinutesTotal != 0 || len(s.OpenGuesses) != 0 {
				t.Errorf("after: total %d, %d open guesses; want 0 and 0", s.GuessMinutesTotal, len(s.OpenGuesses))
			}
		})
	}

	refusals := []struct {
		name    string
		how     int
		session string
		want    string
	}{
		{"unknown_session", inTerminalDesk, "not-a-desk", state.RefuseNotDesk},
		{"not_under_desk", outsideDesk, deskSession, state.RefuseNotUnderDesk},
		{"no_terminal", inSetsidDesk, deskSession, state.RefuseNoTerminal},
	}
	for _, c := range refusals {
		t.Run(c.name, func(t *testing.T) {
			dir := newStateFrom(t, "order")
			readRulingSnap(t, dir, at(t, 0)) // writes snapshot.json and usage.json, so the folder has bytes to compare
			var before map[string]string
			r := recordRuling(t, c.how, c.session, dir, at(t, 20), "north-q1", "yes", func() { before = folder(t, dir) })
			if r.code != 2 || !strings.Contains(r.stderr, c.want) {
				t.Fatalf("exit %d, want 2 with %q; stderr: %s", r.code, c.want, r.stderr)
			}
			requireSnapshot(t, before)
			sameFolder(t, before, folder(t, dir))
		})
	}
}
