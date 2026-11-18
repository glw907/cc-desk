package sessions

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestParsePulls(t *testing.T) {
	merged := time.Date(2026, 11, 18, 7, 40, 7, 0, time.UTC)
	tests := []struct {
		name string
		in   string
		want []Pull
	}{
		{"merged with a passing run",
			`[{"headRefOid":"1c128e6f","mergeCommit":{"oid":"3b65cda9"},"mergedAt":"2026-11-18T07:40:07Z","number":5,"state":"MERGED","statusCheckRollup":[{"__typename":"CheckRun","completedAt":"2026-11-18T07:39:42Z","conclusion":"SUCCESS","name":"gate","startedAt":"2026-11-18T07:39:06Z","status":"COMPLETED","workflowName":"gate"}]}]`,
			[]Pull{{Number: 5, State: "MERGED", Head: "1c128e6f", Merged: &merged, MergeCommit: "3b65cda9", Checks: []Check{{"gate", Pass}}}}},
		{"open with a passing run",
			`[{"headRefOid":"abc","mergeCommit":null,"mergedAt":null,"number":12,"state":"OPEN","statusCheckRollup":[{"__typename":"CheckRun","conclusion":"SUCCESS","name":"gate","status":"COMPLETED","workflowName":"gate"}]}]`,
			[]Pull{{Number: 12, State: "OPEN", Head: "abc", Checks: []Check{{"gate", Pass}}}}},
		{"no checks", `[{"number":1,"statusCheckRollup":[]}]`, []Pull{{Number: 1}}},
		{"no pull request", `[]`, nil},
		{"a status context",
			`[{"number":2,"state":"OPEN","statusCheckRollup":[{"__typename":"StatusContext","context":"ci/build","state":"FAILURE"},{"__typename":"CheckRun","name":"lint","status":"IN_PROGRESS","conclusion":""}]}]`,
			[]Pull{{Number: 2, State: "OPEN", Checks: []Check{{"ci/build", Fail}, {"lint", Pending}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePulls([]byte(tt.in))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
	if _, err := ParsePulls([]byte("nope")); err == nil {
		t.Error("garbage parsed without an error")
	}
}

func TestReadPulls(t *testing.T) {
	script := func(t *testing.T, body string) string {
		t.Helper()
		p := filepath.Join(t.TempDir(), "gh")
		if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	tests := []struct {
		name        string
		gh          string
		unreachable bool
		pulls       int
	}{
		{"an answer", script(t, `echo '[{"number":3,"state":"OPEN"}]'`), false, 1},
		{"no pull request", script(t, `echo '[]'`), false, 0},
		{"a non-zero exit", script(t, `echo 'GraphQL: Could not resolve' >&2; exit 1`), true, 0},
		{"no such command", "/no/such/gh", true, 0},
		{"output that is not a list", script(t, `echo nope`), true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadPulls(context.Background(), tt.gh, "o/r", "b")
			if got.Unreachable != tt.unreachable || len(got.Pulls) != tt.pulls {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func TestChoose(t *testing.T) {
	r := PullRead{Pulls: []Pull{{Number: 1}, {Number: 2}}}
	if p, ok := r.Choose(2); !ok || p.Number != 2 {
		t.Errorf("named number: %+v %v", p, ok)
	}
	if p, ok := r.Choose(9); !ok || p.Number != 1 {
		t.Errorf("fallback: %+v %v", p, ok)
	}
	if _, ok := (PullRead{}).Choose(1); ok {
		t.Error("empty read chose a pull request")
	}
}

func TestBucket(t *testing.T) {
	tests := []struct {
		state, status, conclusion string
		want                      CheckState
	}{
		{"SUCCESS", "", "", Pass},
		{"", "COMPLETED", "SUCCESS", Pass},
		{"", "COMPLETED", "SKIPPED", Skipping},
		{"", "COMPLETED", "NEUTRAL", Skipping},
		{"ERROR", "", "", Fail},
		{"FAILURE", "", "", Fail},
		{"", "COMPLETED", "TIMED_OUT", Fail},
		{"", "COMPLETED", "ACTION_REQUIRED", Fail},
		{"", "COMPLETED", "CANCELLED", Cancel},
		{"EXPECTED", "", "", Pending},
		{"PENDING", "", "", Pending},
		{"", "REQUESTED", "", Pending},
		{"", "WAITING", "", Pending},
		{"", "QUEUED", "", Pending},
		{"", "IN_PROGRESS", "", Pending},
		{"", "COMPLETED", "STALE", Pending},
		{"", "", "", Pending},
	}
	for _, tt := range tests {
		t.Run(tt.state+"/"+tt.status+"/"+tt.conclusion, func(t *testing.T) {
			if got := bucket(tt.state, tt.status, tt.conclusion); got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}
