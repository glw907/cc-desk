package state

import (
	"strings"
	"testing"
)

func closedState() (*State, *ReleaseFile) {
	s, rf := ruleState()
	s.Rulings = []Ruling{{Format: Format, ID: "r-0001", Time: at(9, 0), Desk: "d", Words: "Close it."}}
	return s, rf
}

// startedPass adds a pass in state st that started at 8:00.
func startedPass(rf *ReleaseFile, name string, st PassState, after ...string) *Pass {
	p := addPass(rf, name, "r1", 60, st, after...)
	t := at(8, 0)
	p.Started = &t
	return p
}

func TestClosePass(t *testing.T) {
	t.Run("a building pass", func(t *testing.T) {
		s, rf := closedState()
		startedPass(rf, "p1", Building)
		if err := ClosePass("alpha", "p1", "r-0001")(s, at(9, 30)); err != nil {
			t.Fatal(err)
		}
		p := &rf.Passes[0]
		if p.State != Done || p.Closed == nil || p.Closed.Ruling != "r-0001" || !p.Closed.Time.Equal(at(9, 30)) {
			t.Errorf("pass = %+v", p)
		}
		if p.Finished == nil || !p.Finished.Equal(at(9, 30)) || p.ActualMinutes == nil || *p.ActualMinutes != 90 {
			t.Errorf("finished %v, actual %v", p.Finished, p.ActualMinutes)
		}
		if !rf.Dirty || !s.passEvent {
			t.Error("the close was not marked for writing")
		}
	})

	t.Run("a stopped pass closes its stop", func(t *testing.T) {
		s, rf := closedState()
		p := startedPass(rf, "p1", Stopped)
		p.Stops = []StopInterval{{Start: at(8, 40)}}
		p.StopReason, p.StopDetail = StopScope, "d"
		if err := ClosePass("alpha", "p1", "r-0001")(s, at(9, 30)); err != nil {
			t.Fatal(err)
		}
		// 90 minutes since the start, less the 50 stopped.
		if p.State != Done || *p.ActualMinutes != 40 || p.Stops[0].End == nil || !p.Stops[0].End.Equal(at(9, 30)) {
			t.Errorf("pass = %+v", p)
		}
		if p.StopReason != "" || p.StopDetail != "" {
			t.Errorf("stop reason %q, detail %q are still set", p.StopReason, p.StopDetail)
		}
	})

	for _, tt := range []struct {
		name  string
		st    PassState
		extra string
	}{
		{"a pass not started", NotStarted, "is not_started and has no pull request to close"},
		{"a waiting pass", Waiting, "is waiting and has no pull request to close"},
		{"a done pass", Done, "is already done"},
	} {
		t.Run("refuses "+tt.name, func(t *testing.T) {
			s, rf := closedState()
			p := addPass(rf, "p1", "r1", 60, tt.st)
			if tt.st == Done {
				st, fin, n := at(8, 0), at(9, 0), 60
				p.Started, p.Finished, p.ActualMinutes = &st, &fin, &n
			}
			err := ClosePass("alpha", "p1", "r-0001")(s, at(9, 30))
			if !isRefusal(err) || !strings.Contains(err.Error(), tt.extra) {
				t.Fatalf("error = %v, want a refusal with %q", err, tt.extra)
			}
			if p.Closed != nil {
				t.Error("a refused close marked the pass")
			}
		})
	}

	t.Run("an unknown ruling is a bad request", func(t *testing.T) {
		s, rf := closedState()
		startedPass(rf, "p1", Building)
		if err := ClosePass("alpha", "p1", "r-0009")(s, at(9, 30)); err == nil || isRefusal(err) {
			t.Errorf("error = %v, want a bad request", err)
		}
	})
}

// closedBefore is a release whose p1 was closed and whose p2 follows it.
func closedBefore() (*State, *ReleaseFile) {
	s, rf := closedState()
	startedPass(rf, "p1", Building)
	if err := ClosePass("alpha", "p1", "r-0001")(s, at(9, 30)); err != nil {
		panic(err)
	}
	addPass(rf, "p2", "r1", 30, NotStarted, "p1")
	for i := range rf.Passes {
		rf.Passes[i].Summary = "s"
	}
	return s, rf
}

func TestBreaksClosed(t *testing.T) {
	s, rf := closedBefore()
	got := s.Breaks(rf, &rf.Passes[1], "")
	if len(got) != 1 || got[0].Name != BreakClosed || got[0].Detail != "pass p1 was closed" {
		t.Fatalf("breaks = %+v", got)
	}

	t.Run("a merged pass before it breaks nothing", func(t *testing.T) {
		s, rf := closedBefore()
		rf.Passes[0].Closed = nil
		if got := s.Breaks(rf, &rf.Passes[1], ""); len(got) != 0 {
			t.Errorf("breaks = %+v", got)
		}
	})
}

func TestValidateClosed(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rf *ReleaseFile)
		wantErr string
	}{
		{name: "a closed done pass reads"},
		{
			name: "closed on a building pass", wantErr: "is building but is marked closed",
			setup: func(rf *ReleaseFile) {
				rf.Passes[0].State, rf.Passes[0].Finished, rf.Passes[0].ActualMinutes = Building, nil, nil
			},
		},
		{
			name: "closed with no ruling", wantErr: `closed needs a "ruling" and a "time"`,
			setup: func(rf *ReleaseFile) { rf.Passes[0].Closed.Ruling = "" },
		},
		{
			name: "closed with a merged pull request", wantErr: "its pull_request is merged",
			setup: func(rf *ReleaseFile) {
				m := at(9, 0)
				rf.Passes[0].PullRequest = &PullRequest{Number: 3, Merged: &m}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, rf := closedBefore()
			if tt.setup != nil {
				tt.setup(rf)
			}
			err := validateRelease(&rf.Release)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatal(err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}

	t.Run("the ruling must be in the log", func(t *testing.T) {
		s, _ := closedBefore()
		s.Rulings = nil
		err := Check(s)
		if err == nil || !strings.Contains(err.Error(), `closed by ruling "r-0001"`) {
			t.Errorf("error = %v", err)
		}
	})
}
