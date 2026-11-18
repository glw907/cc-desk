package usage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

const liveShape = `{"type":"result","subtype":"success","is_error":false,"num_turns":0,"total_cost_usd":0,` +
	`"result":"You are currently using your subscription to power your Claude Code usage\nCurrent week (all models): 89% used · resets Nov 20, 10:59pm (America/Anchorage)\nCurrent week (Fable): 83% used · resets Nov 20, 10:59pm (America/Anchorage)\n"}`

const builtShape = `{"is_error":false,"result":"Current week (Fable): 83% used · resets Nov 20, 10:59pm (America/Anchorage)\n` +
	`What's contributing\n96% of your usage came from one session\nCurrent week (all models): 41% used · resets Nov 21, 1am (America/Anchorage)\n"}`

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		percent int
		reset   string
		err     string
	}{
		{"live shape", liveShape, 89, "Nov 20, 10:59pm (America/Anchorage)", ""},
		{"earlier percents are skipped", builtShape, 41, "Nov 21, 1am (America/Anchorage)", ""},
		{"bad JSON", `{"result":`, 0, "", "not JSON"},
		{"is_error", `{"is_error":true,"result":"Current week (all models): 10% used · resets x"}`, 0, "", "reports an error"},
		{"no matching line", `{"is_error":false,"result":"Current week (Fable): 83% used · resets x"}`, 0, "", "no \"Current week (all models)\" line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			percent, reset, err := Parse([]byte(tt.raw))
			if tt.err != "" {
				if err == nil || !strings.Contains(err.Error(), tt.err) {
					t.Fatalf("error = %v, want it to hold %q", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if percent != tt.percent || reset != tt.reset {
				t.Errorf("got %d, %q; want %d, %q", percent, reset, tt.percent, tt.reset)
			}
		})
	}
}

func script(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "stand-in.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRead(t *testing.T) {
	t.Run("returns its output", func(t *testing.T) {
		out, err := Read(context.Background(), script(t, "printf hello"))
		if err != nil || string(out) != "hello" {
			t.Errorf("out %q err %v", out, err)
		}
	})
	t.Run("a non-zero exit is an error", func(t *testing.T) {
		if _, err := Read(context.Background(), script(t, "exit 3")); err == nil {
			t.Error("a failing command returned no error")
		}
	})
	t.Run("a missing command is an error", func(t *testing.T) {
		if _, err := Read(context.Background(), filepath.Join(t.TempDir(), "nope")); err == nil {
			t.Error("a missing command returned no error")
		}
	})
	t.Run("a timeout returns before the command ends", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		start := time.Now()
		_, err := Read(ctx, script(t, "sleep 4"))
		if err == nil || !strings.Contains(err.Error(), "timed out") {
			t.Errorf("error = %v, want a timeout", err)
		}
		if d := time.Since(start); d >= 4*time.Second {
			t.Errorf("took %s, want under 4s", d)
		}
	})
}

func TestFresh(t *testing.T) {
	now := time.Date(2026, 11, 20, 23, 0, 0, 0, time.UTC)
	reading := func(at time.Time) string {
		return `{"format":1,"percent":40,"reset_text":"x","read_at":"` + at.Format(time.RFC3339) + `"}`
	}
	tests := []struct {
		name string
		file string // "" for none
		want bool
	}{
		{"absent", "", false},
		{"broken", `{"format":1,`, false},
		{"no read time", `{"format":1,"percent":40,"reset_text":"x"}`, false},
		{"4 minutes old", reading(now.Add(-4 * time.Minute)), true},
		{"5 minutes old", reading(now.Add(-5 * time.Minute)), false},
		{"in the future", reading(now.Add(time.Minute)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.file != "" {
				if err := os.WriteFile(filepath.Join(dir, state.UsageFile), []byte(tt.file), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if _, got := Cached(dir, now); got != tt.want {
				t.Errorf("fresh = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTake(t *testing.T) {
	now := time.Date(2026, 11, 20, 23, 0, 0, 0, time.UTC)
	good := script(t, "cat <<'EOF'\n"+liveShape+"\nEOF")
	bad := script(t, "exit 1")

	t.Run("a fresh cache runs nothing", func(t *testing.T) {
		dir := t.TempDir()
		cached := `{"format":1,"percent":12,"reset_text":"x","read_at":"2026-11-20T22:59:00Z"}`
		if err := os.WriteFile(filepath.Join(dir, state.UsageFile), []byte(cached), 0o644); err != nil {
			t.Fatal(err)
		}
		u, raw, took := Take(context.Background(), dir, bad, now, false)
		if took || raw != nil || u.Percent != 12 {
			t.Errorf("u %+v raw %q took %v", u, raw, took)
		}
		u, _, took = Take(context.Background(), dir, good, now, true)
		if !took || u.Percent != 89 || !u.ReadAt.Equal(now) {
			t.Errorf("forced: u %+v took %v", u, took)
		}
	})
	t.Run("a failed run is an error reading", func(t *testing.T) {
		u, _, took := Take(context.Background(), t.TempDir(), bad, now, false)
		if !took || u.Error == "" || u.Percent != 0 || u.ResetText != "" {
			t.Errorf("u %+v took %v", u, took)
		}
	})
	t.Run("an unreadable result is an error reading with its raw bytes", func(t *testing.T) {
		u, raw, _ := Take(context.Background(), t.TempDir(), script(t, "printf nope"), now, false)
		if u.Error == "" || string(raw) != "nope" {
			t.Errorf("u %+v raw %q", u, raw)
		}
	})
}
