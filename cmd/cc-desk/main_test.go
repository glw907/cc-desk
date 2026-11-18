package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

func TestVersion(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "cc-desk version 0.1.0\n"; got != want {
		t.Errorf("--version wrote %q, want %q", got, want)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"refusal", state.Refuse("already done"), 2},
		{"wrapped refusal", fmt.Errorf("pass: %w", state.Refuse("x")), 2},
		{"plain error", errors.New("broken"), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCode(tt.err); got != tt.want {
				t.Errorf("exitCode = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClock(t *testing.T) {
	f := rootFlags{now: "2026-11-20T14:00:00-09:00"}
	got, err := f.clock()
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 11, 20, 23, 0, 0, 0, time.UTC); !got.Equal(want) || got.Location() != time.UTC {
		t.Errorf("clock = %v, want %v in UTC", got, want)
	}
	bad := rootFlags{now: "yesterday"}
	if _, err := bad.clock(); err == nil {
		t.Error("clock accepted a bad time")
	}
}
