package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/state"
)

// launch starts the binary and returns a function that waits for it. It
// runs from the test goroutine, so a failure to start fails the test.
func launch(t *testing.T, usageCommand, dir, now string, args ...string) func() result {
	t.Helper()
	full := append(globalArgsWith(usageCommand, dir, now), args...)
	cmd := exec.Command(binary, full...)
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %v: %v", args, err)
	}
	return func() result {
		code := 0
		if err := cmd.Wait(); err != nil {
			ee, ok := errors.AsType[*exec.ExitError](err)
			if !ok {
				return result{out.String(), err.Error(), -1}
			}
			code = ee.ExitCode()
		}
		return result{out.String(), errOut.String(), code}
	}
}

func globalArgsWith(usageCommand, dir, now string) []string {
	return []string{"--state-dir", dir, "--now", now, "--usage-command", usageCommand}
}

// TestTwoWriters runs twenty verbs at once against one folder: seven
// finishes, seven stops, and six rulings from a desk. Each takes a usage
// reading outside the lock, so the 15-second stand-in runs in parallel.
func TestTwoWriters(t *testing.T) {
	const deskSession = "e2e-writers-desk"
	dir := newStateFrom(t, "writers")
	standin := newStandIn(t, 15, usageResult(40))
	now := workedClock

	// The ruling verbs run from a desk with a terminal, each in the
	// ( ... ) & form with its own result files.
	out := t.TempDir()
	var script strings.Builder
	questions := []struct{ project, round, question string }{
		{"east", "e1", "east-q1"}, {"east", "e1", "east-q2"}, {"east", "e1", "east-q3"},
		{"west", "w1", "west-q1"}, {"west", "w1", "west-q2"}, {"west", "w1", "west-q3"},
	}
	for i, q := range questions {
		args := append(globalArgsWith(standin.path, dir, now),
			"ruling", "record", "--project", q.project, "--round", q.round, "--question", q.question, "--option", "yes",
			"--", "Words for "+q.question+".")
		fmt.Fprintf(&script, "( setsid -w \"$BIN\" %s > \"$OUT/o%d\" 2> \"$OUT/e%d\"; echo $? > \"$OUT/c%d\" ) &\n",
			shellWords(args), i, i, i)
	}
	pid, release, wait := startDesk(t, true, []string{"BIN=" + binary, "OUT=" + out, "CLAUDE_CODE_SESSION_ID=" + deskSession}, script.String())
	entry, err := desk.Entry(deskSession, pid)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state.Desks{Format: state.Format, Desks: []state.Desk{entry}})
	if err != nil {
		t.Fatal(err)
	}
	// Written before any verb starts, so none reads a half-written file.
	writeIn(t, dir, "desks.json", string(data)+"\n")

	start := time.Now()
	var waits []func() result
	for _, p := range []string{"e-1", "e-2", "e-3", "e-4"} {
		waits = append(waits, launch(t, standin.path, dir, now, "pass", "finish", "east", p))
	}
	for _, p := range []string{"w-1", "w-2", "w-3"} {
		waits = append(waits, launch(t, standin.path, dir, now, "pass", "finish", "west", p))
	}
	for _, p := range []string{"e-5", "e-6", "e-7"} {
		waits = append(waits, launch(t, standin.path, dir, now, "pass", "stop", "east", p, "--reason", "scope"))
	}
	for _, p := range []string{"w-4", "w-5", "w-6", "w-7"} {
		waits = append(waits, launch(t, standin.path, dir, now, "pass", "stop", "west", p, "--reason", "scope"))
	}
	release()

	var wg sync.WaitGroup
	results := make([]result, len(waits))
	for i, w := range waits {
		wg.Go(func() { results[i] = w() })
	}
	wg.Wait()
	wait()
	elapsed := time.Since(start)

	for i, r := range results {
		if r.code != 0 {
			t.Errorf("verb %d exit %d: %s", i, r.code, r.stderr)
		}
	}
	for i := range questions {
		code, err := strconv.Atoi(strings.TrimSpace(readFile(t, filepath.Join(out, fmt.Sprintf("c%d", i)))))
		if err != nil || code != 0 {
			t.Errorf("ruling %d: exit %q: %s", i, readFile(t, filepath.Join(out, fmt.Sprintf("c%d", i))), readFile(t, filepath.Join(out, fmt.Sprintf("e%d", i))))
		}
	}

	times := standin.callTimes(t)
	if len(times) != 20 {
		t.Fatalf("stand-in ran %d times, want 20", len(times))
	}
	if spread := slices.Max(times) - slices.Min(times); spread > 5 {
		t.Errorf("stand-in calls spread over %.1f seconds, want within 5", spread)
	}
	if elapsed >= 40*time.Second {
		t.Errorf("twenty verbs took %v, want under 40 seconds", elapsed)
	}

	for project, passes := range map[string]map[string]string{
		"east": {"e-1": "done", "e-2": "done", "e-3": "done", "e-4": "done", "e-5": "stopped", "e-6": "stopped", "e-7": "stopped"},
		"west": {"w-1": "done", "w-2": "done", "w-3": "done", "w-4": "stopped", "w-5": "stopped", "w-6": "stopped", "w-7": "stopped"},
	} {
		var rel struct {
			Passes []struct {
				Name  string `json:"name"`
				State string `json:"state"`
			} `json:"passes"`
			Questions []struct {
				ID     string `json:"id"`
				Answer *struct {
					Ruling string `json:"ruling"`
				} `json:"answer"`
			} `json:"questions"`
		}
		if err := json.Unmarshal([]byte(readFile(t, filepath.Join(dir, "releases", project+".json"))), &rel); err != nil {
			t.Fatal(err)
		}
		for _, p := range rel.Passes {
			if p.State != passes[p.Name] {
				t.Errorf("%s is %s, want %s", p.Name, p.State, passes[p.Name])
			}
		}
		for _, q := range rel.Questions {
			if q.Answer == nil || !strings.HasPrefix(q.Answer.Ruling, "r-") {
				t.Errorf("%s has no answer naming a ruling", q.ID)
			}
		}
	}

	rulings := nonEmptyLines(readFile(t, filepath.Join(dir, "rulings.jsonl")))
	ids := map[string]bool{}
	for _, l := range rulings {
		var r state.Ruling
		if err := json.Unmarshal([]byte(l), &r); err != nil {
			t.Fatalf("rulings.jsonl line %q: %v", l, err)
		}
		ids[r.ID] = true
	}
	if len(rulings) != 6 || len(ids) != 6 {
		t.Errorf("rulings.jsonl has %d lines and %d distinct ids, want 6 and 6", len(rulings), len(ids))
	}
	if n := len(nonEmptyLines(readFile(t, filepath.Join(dir, "forecasts.jsonl")))); n != 14 {
		t.Errorf("forecasts.jsonl has %d lines, want 14", n)
	}

	// The snapshot verb finds the cached reading, so it never calls the
	// stand-in and prints the bytes the last commit wrote.
	onDisk := readFile(t, filepath.Join(dir, "snapshot.json"))
	r := runWith(t, standin.path, dir, now, "snapshot")
	if r.code != 0 {
		t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
	}
	if r.stdout != onDisk {
		t.Error("the snapshot verb's output differs from snapshot.json as the writers left it")
	}
	if n := len(standin.callTimes(t)); n != 20 {
		t.Errorf("stand-in log holds %d lines after the snapshot verb, want 20", n)
	}
}

func nonEmptyLines(s string) []string {
	var out []string
	for l := range strings.SplitSeq(s, "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
