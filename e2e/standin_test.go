package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// standIn is a per-test stand-in for the usage command. It prints the
// result file after an optional delay, and logs one line per call.
type standIn struct {
	path, resultFile, logFile string
}

// usageResult is a /usage result in the live shape with the given percent
// on the "Current week (all models)" line.
func usageResult(percent int) string {
	return fmt.Sprintf(`{"type":"result","subtype":"success","is_error":false,"num_turns":0,"total_cost_usd":0,`+
		`"result":"You are currently using your subscription to power your Claude Code usage\nCurrent week (all models): %d%% used · resets Nov 20, 10:59pm (America/Anchorage)\nCurrent week (Fable): 30%% used · resets Nov 20, 10:59pm (America/Anchorage)\n"}`+"\n", percent)
}

// newStandIn writes a stand-in that sleeps delaySeconds, then prints result.
func newStandIn(t *testing.T, delaySeconds int, result string) *standIn {
	t.Helper()
	dir := t.TempDir()
	s := &standIn{
		path:       filepath.Join(dir, "usage.sh"),
		resultFile: filepath.Join(dir, "result.json"),
		logFile:    filepath.Join(dir, "calls.log"),
	}
	script := fmt.Sprintf("#!/bin/sh\necho call >> %q\nsleep %d\ncat %q\n", s.logFile, delaySeconds, s.resultFile)
	if err := os.WriteFile(s.path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	s.set(t, result)
	return s
}

// set replaces what the stand-in prints from the next call on.
func (s *standIn) set(t *testing.T, result string) {
	t.Helper()
	if err := os.WriteFile(s.resultFile, []byte(result), 0o644); err != nil {
		t.Fatal(err)
	}
}

// calls returns how many times the stand-in has run.
func (s *standIn) calls(t *testing.T) int {
	t.Helper()
	b, err := os.ReadFile(s.logFile)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(b), "call\n")
}

// run is run with this stand-in as the usage command.
func (s *standIn) run(t *testing.T, dir, now string, args ...string) result {
	t.Helper()
	return runWith(t, s.path, dir, now, args...)
}
