package state

import (
	"strings"
	"testing"
)

const designWithContext = `{"format":1,
  "design":"docs/specs/d.md",
  "context":["First paragraph.","Second paragraph."],
  "calls":[{"call":"c","chosen":"x","other":"y","why":"w"}],
  "passes":[{"name":"p2","summary":"More.","round":"r1","planned_minutes":30,"state":"not_started","after":["p1"],"questions":["q1"]}],
  "questions":[{"id":"q1","text":"Which?","options":["a","b"],"recommended":"b"}]}`

func TestDesignRoundContext(t *testing.T) {
	s := loaded(t)
	s.Releases[0].Rounds[0].DesignWritten = false
	d, err := DecodeDesign([]byte(designWithContext))
	if err != nil {
		t.Fatal(err)
	}
	if err := DesignRound("alpha", "r1", d)(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	rd := s.Releases[0].Rounds[0]
	if rd.Design != "docs/specs/d.md" || len(rd.Context) != 2 || len(rd.Calls) != 1 || rd.Calls[0].Chosen != "x" {
		t.Errorf("round after design = %+v", rd)
	}
}

func TestDecodeDesignNotes(t *testing.T) {
	tests := []struct {
		name, from, to, want string
	}{
		{"empty context", `"First paragraph."`, `" "`, "context[0] is empty"},
		{"empty call part", `"why":"w"`, `"why":""`, `calls[0] is missing "why"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeDesign([]byte(strings.Replace(designWithContext, tt.from, tt.to, 1)))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one containing %q", err, tt.want)
			}
		})
	}
}

func TestStartConversation(t *testing.T) {
	s := loaded(t)
	s.Releases[0].Rounds[0].DesignWritten = false
	if err := StartConversation("alpha", "r1", "desk-1")(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	if len(s.Desks.Conversations) != 1 || !s.desksDirty || s.Desks.Conversations[0].Desk != "desk-1" {
		t.Fatalf("conversations = %+v, dirty %v", s.Desks.Conversations, s.desksDirty)
	}
	if err := StartConversation("alpha", "r1", "desk-2")(s, at(9, 5)); !isRefusal(err) {
		t.Errorf("a second conversation: error = %v, want a refusal", err)
	}
	for _, tt := range []struct{ project, round, desk string }{
		{"nope", "r1", "d"}, {"alpha", "r9", "d"}, {"alpha", "r1", ""},
	} {
		if err := StartConversation(tt.project, tt.round, tt.desk)(s, at(9, 6)); err == nil {
			t.Errorf("StartConversation(%+v) succeeded", tt)
		}
	}
}

func TestStartConversationDesigned(t *testing.T) {
	s := loaded(t)
	if err := StartConversation("alpha", "r1", "desk-1")(s, at(9, 0)); !isRefusal(err) {
		t.Errorf("a designed round: error = %v, want a refusal", err)
	}
}

func TestDesignRoundEndsConversation(t *testing.T) {
	s := loaded(t)
	s.Releases[0].Rounds[0].DesignWritten = false
	if err := StartConversation("alpha", "r1", "desk-1")(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	s.desksDirty = false
	d, err := DecodeDesign([]byte(design))
	if err != nil {
		t.Fatal(err)
	}
	if err := DesignRound("alpha", "r1", d)(s, at(9, 10)); err != nil {
		t.Fatal(err)
	}
	if len(s.Desks.Conversations) != 0 || !s.desksDirty {
		t.Errorf("after design: conversations %+v, dirty %v", s.Desks.Conversations, s.desksDirty)
	}
}

func TestLoadDesksConversations(t *testing.T) {
	tests := []struct{ name, desks, want string }{
		{"old file", `{"format":1,"desks":[{"session_id":"d","pid":1,"started":"2026-11-20T09:00:00Z"}]}`, ""},
		{"one", `{"format":1,"conversations":[{"project":"alpha","round":"r1","desk":"d","started":"2026-11-20T09:00:00Z"}]}`, ""},
		{"missing part", `{"format":1,"conversations":[{"project":"alpha","round":"r1","started":"2026-11-20T09:00:00Z"}]}`, "conversation 1 is missing"},
		{"twice", `{"format":1,"conversations":[` +
			`{"project":"alpha","round":"r1","desk":"d","started":"2026-11-20T09:00:00Z"},` +
			`{"project":"alpha","round":"r1","desk":"e","started":"2026-11-20T09:01:00Z"}]}`, "two conversations"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "releases/alpha.json", startRelease)
			writeFile(t, dir, "desks.json", tt.desks)
			_, err := Load(dir)
			switch {
			case tt.want == "" && err != nil:
				t.Errorf("Load: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Errorf("Load error = %v, want one containing %q", err, tt.want)
			}
		})
	}
}
