package state

import (
	"strings"
	"testing"
	"time"
)

func TestRefusalRecord(t *testing.T) {
	good := Refusal{Time: at(9, 40), Tool: "Bash", Target: "git stash drop stash@{0}", Reason: "dangerous"}
	tests := []struct {
		name      string
		sessionID string
		refusal   Refusal
		setup     func(rf *ReleaseFile)
		wantFound bool
		wantErr   bool
	}{
		{name: "a pass's session", sessionID: "sess-x", refusal: good, wantFound: true},
		{name: "a conductor's session", sessionID: "sess-c", refusal: good, wantFound: true},
		{name: "an empty reason", sessionID: "sess-x", refusal: Refusal{Time: at(9, 40), Tool: "Bash"}, wantFound: true},
		{name: "an unknown session", sessionID: "nope", refusal: good},
		{name: "an empty session id", refusal: good},
		{
			name: "an ended record is not open", sessionID: "sess-x", refusal: good,
			setup: func(rf *ReleaseFile) { e := at(9, 35); rf.Passes[0].Session.Ended = &e },
		},
		{name: "no tool", sessionID: "sess-x", refusal: Refusal{Time: at(9, 40)}, wantErr: true},
		{name: "no time", sessionID: "sess-x", refusal: Refusal{Tool: "Bash"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := promptState()
			if tt.setup != nil {
				tt.setup(rf)
			}
			var found bool
			err := RecordRefusal(tt.sessionID, tt.refusal, &found)(s, at(9, 40))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error %v", err, tt.wantErr)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound && (!s.noSettle || !rf.Dirty) {
				t.Errorf("noSettle = %v, dirty = %v, want both true", s.noSettle, rf.Dirty)
			}
			if tt.wantErr && s.noSettle {
				t.Error("a refused refusal set noSettle")
			}
		})
	}

	t.Run("it appends to the right record", func(t *testing.T) {
		s, rf := promptState()
		var found bool
		if err := RecordRefusal("sess-x", good, &found)(s, at(9, 40)); err != nil {
			t.Fatal(err)
		}
		if n := len(rf.Passes[0].Session.Refusals); n != 1 || rf.Passes[0].Session.Refusals[0] != good {
			t.Errorf("pass refusals = %+v, want [%+v]", rf.Passes[0].Session.Refusals, good)
		}
		if n := len(rf.Conductor.Refusals); n != 0 {
			t.Errorf("conductor refusals = %d, want 0", n)
		}
	})
}

func TestRelaunchKeepsRefusals(t *testing.T) {
	refused := strings.Replace(goodSession, `"launch_tries":1`,
		`"launch_tries":2,"refusals":[{"time":"2026-11-20T10:02:00Z","tool":"Bash","target":"git stash drop stash@{0}","reason":"dangerous"}]`, 1)
	for _, kind := range []LaunchKind{LaunchAgain, LaunchRelaunch} {
		dir, _ := loadDir(t, sessionRelease("building", refused, ""))
		ev := RecordLaunch("p", "pass:x", Launch{Kind: kind, Replaces: "ab12", Session: bareSession("p-one", "/w/one"), Cause: "idle"})
		if _, err := Commit(dir, testClock, time.Second, ev, render); err != nil {
			t.Fatal(err)
		}
		s, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		got := s.Releases[0].Passes[0].Session.Refusals
		if len(got) != 1 || got[0].Target != "git stash drop stash@{0}" || got[0].Reason != "dangerous" {
			t.Errorf("launch kind %d: refusals = %+v, want the one from the old record", kind, got)
		}
	}
}
