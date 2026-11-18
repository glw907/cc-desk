package state

import (
	"slices"
	"strings"
	"testing"
	"time"
)

// promptState is deskState with a pass x stopped on a permission prompt, its
// session raising one unanswered Bash prompt, and a conductor session.
func promptState() (*State, *ReleaseFile) {
	s, rf := deskState()
	t0, t1 := at(9, 0), at(9, 30)
	rf.Passes = []Pass{{
		Name: "x", Round: "r1", PlannedMinutes: 10, State: Stopped, Started: &t0,
		StopReason: StopPermissionPrompt, Stops: []StopInterval{{Start: t1}},
		Session: &Session{
			Name: "alpha-x", Kind: KindPass, ShortID: "ab12", SessionID: "sess-x", Worktree: "/w/x", Branch: "x",
			Prompts: []Prompt{{Time: at(9, 29), Tool: "Bash", Command: "go test ./...", SuggestedRule: "Bash(go test *)"}},
		},
	}}
	rf.Conductor = &Session{
		Name: "alpha-conductor", Kind: KindConductor, ShortID: "cd34", SessionID: "sess-c", Worktree: "/w/c", Branch: "c",
		Prompts: []Prompt{{Time: at(9, 29), Tool: "Bash", Command: "ls"}},
	}
	return s, rf
}

func TestRecordPrompt(t *testing.T) {
	good := Prompt{Time: at(9, 40), Tool: "Bash", Command: "make"}
	tests := []struct {
		name      string
		sessionID string
		prompt    Prompt
		setup     func(rf *ReleaseFile)
		wantFound bool
		wantErr   bool
	}{
		{name: "a pass's session", sessionID: "sess-x", prompt: good, wantFound: true},
		{name: "a conductor's session", sessionID: "sess-c", prompt: good, wantFound: true},
		{name: "an unknown session", sessionID: "nope", prompt: good},
		{name: "an empty session id", prompt: good},
		{
			name: "an ended record is not open", sessionID: "sess-x", prompt: good,
			setup: func(rf *ReleaseFile) { e := at(9, 35); rf.Passes[0].Session.Ended = &e },
		},
		{name: "no tool", sessionID: "sess-x", prompt: Prompt{Time: at(9, 40)}, wantErr: true},
		{name: "no time", sessionID: "sess-x", prompt: Prompt{Tool: "Bash"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := promptState()
			if tt.setup != nil {
				tt.setup(rf)
			}
			var found bool
			err := RecordPrompt(tt.sessionID, tt.prompt, &found)(s, at(9, 40))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error %v", err, tt.wantErr)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound {
				if !s.noSettle || !rf.Dirty {
					t.Errorf("noSettle = %v, dirty = %v, want both true", s.noSettle, rf.Dirty)
				}
			}
			if tt.wantErr && s.noSettle {
				t.Error("a refused prompt set noSettle")
			}
		})
	}

	t.Run("it appends to the right record", func(t *testing.T) {
		s, rf := promptState()
		var found bool
		if err := RecordPrompt("sess-x", good, &found)(s, at(9, 40)); err != nil {
			t.Fatal(err)
		}
		if n := len(rf.Passes[0].Session.Prompts); n != 2 {
			t.Errorf("pass prompts = %d, want 2", n)
		}
		if n := len(rf.Conductor.Prompts); n != 1 {
			t.Errorf("conductor prompts = %d, want 1", n)
		}
	})
}

func TestAnswerPrompt(t *testing.T) {
	notDesk := func() Caller { c := deskCaller(); c.SessionID = "other"; return c }
	tests := []struct {
		name     string
		caller   func() Caller
		record   string
		answer   string
		rule     string
		setup    func(rf *ReleaseFile)
		refused  string // a substring of the refusal
		wantRule string
	}{
		{name: "a yes for a pass", record: "pass:x", answer: AnswerYes, rule: "Bash(go test ./...)", wantRule: "Bash(go test ./...)"},
		{name: "a yes takes the exact rule of a simple command", record: "pass:x", answer: AnswerYes, wantRule: "Bash(go test ./...)"},
		{name: "a no adds no rule", record: "pass:x", answer: AnswerNo},
		{
			name: "a yes for a conductor", record: "conductor", answer: AnswerYes, wantRule: "Bash(ls)",
			setup: func(rf *ReleaseFile) {
				t := at(9, 30)
				rf.Conductor.StopReason, rf.Conductor.Stopped = StopPermissionPrompt, &t
			},
		},
		{
			name: "a rule already there is not added twice", record: "pass:x", answer: AnswerYes, wantRule: "Bash(go test ./...)",
			setup: func(rf *ReleaseFile) { rf.AllowRules = []string{"Bash(go test ./...)"} },
		},
		{name: "not a desk", caller: notDesk, record: "pass:x", answer: AnswerYes, refused: answerPrefix + reasonNotDesk},
		{
			name: "no terminal", record: "pass:x", answer: AnswerNo, refused: answerPrefix + reasonNoTerminal,
			caller: func() Caller { c := deskCaller(); c.Chain[1].TTY = 0; return c },
		},
		{
			name: "an ended record", record: "pass:x", answer: AnswerNo, refused: "no open session record",
			setup: func(rf *ReleaseFile) { e := at(9, 35); rf.Passes[0].Session.Ended = &e },
		},
		{
			name: "no unanswered prompt", record: "pass:x", answer: AnswerNo, refused: "no unanswered prompt",
			setup: func(rf *ReleaseFile) { rf.Passes[0].Session.Prompts[0].Answer = "no" },
		},
		{
			name: "a building pass", record: "pass:x", answer: AnswerNo, refused: "isn't stopped on a permission prompt",
			setup: func(rf *ReleaseFile) { rf.Passes[0].State, rf.Passes[0].StopReason = Building, "" },
		},
		{
			name: "a pass stopped on sign in", record: "pass:x", answer: AnswerNo, refused: "isn't stopped on a permission prompt",
			setup: func(rf *ReleaseFile) { rf.Passes[0].StopReason = StopSignInExpired },
		},
		{name: "a conductor that isn't stopped", record: "conductor", answer: AnswerNo, refused: "isn't stopped on a permission prompt"},
		{
			name: "a compound command", record: "pass:x", answer: AnswerYes, refused: "needs --rule",
			setup: func(rf *ReleaseFile) { rf.Passes[0].Session.Prompts[0].Command = "make && make test" },
		},
		{
			name: "a redirect", record: "pass:x", answer: AnswerYes, refused: "Bash(go test *)",
			setup: func(rf *ReleaseFile) { rf.Passes[0].Session.Prompts[0].Command = "ls > out.txt" },
		},
		{
			name: "a tool other than Bash", record: "pass:x", answer: AnswerYes, refused: "needs --rule",
			setup: func(rf *ReleaseFile) { rf.Passes[0].Session.Prompts[0].Tool = "Edit" },
		},
		{name: "a malformed rule", record: "pass:x", answer: AnswerYes, rule: "go test", refused: "is not a rule"},
		{name: "a merge rule", record: "pass:x", answer: AnswerYes, rule: "Bash(gh pr merge 5)", refused: "answer no"},
		{name: "a land rule", record: "pass:x", answer: AnswerYes, rule: "Bash(approval-queue land 5)", refused: "answer no"},
		{name: "a bad answer", record: "pass:x", answer: "maybe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := promptState()
			if tt.setup != nil {
				tt.setup(rf)
			}
			caller := deskCaller
			if tt.caller != nil {
				caller = tt.caller
			}
			var out Answered
			err := AnswerPrompt(caller(), "alpha", tt.record, tt.answer, tt.rule, &out)(s, at(10, 0))
			if tt.refused != "" || tt.name == "a bad answer" {
				if err == nil {
					t.Fatal("no error")
				}
				if tt.refused != "" && (!isRefusal(err) || !strings.Contains(err.Error(), tt.refused)) {
					t.Fatalf("error = %v, want a refusal holding %q", err, tt.refused)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if out.Rule != tt.wantRule {
				t.Errorf("rule = %q, want %q", out.Rule, tt.wantRule)
			}
			var sess *Session
			if tt.record == "conductor" {
				sess = rf.Conductor
			} else {
				sess = rf.Passes[0].Session
				if rf.Passes[0].State != Building {
					t.Errorf("pass state = %s, want building", rf.Passes[0].State)
				}
			}
			if got := sess.Prompts[out.Index].Answer; got != tt.answer {
				t.Errorf("answer = %q, want %q", got, tt.answer)
			}
			if n := slices.Index(rf.AllowRules, tt.wantRule); tt.wantRule != "" && n < 0 {
				t.Errorf("allow rules %q lack %q", rf.AllowRules, tt.wantRule)
			}
			if tt.wantRule == "" && len(rf.AllowRules) != 0 {
				t.Errorf("a no changed the allow rules: %q", rf.AllowRules)
			}
			if c := slices.Compact(slices.Clone(rf.AllowRules)); len(c) != len(rf.AllowRules) {
				t.Errorf("allow rules hold a duplicate: %q", rf.AllowRules)
			}
		})
	}

	t.Run("an older unanswered prompt stays", func(t *testing.T) {
		s, rf := promptState()
		sess := rf.Passes[0].Session
		sess.Prompts = append([]Prompt{{Time: at(9, 10), Tool: "Bash", Command: "old"}}, sess.Prompts...)
		var out Answered
		if err := AnswerPrompt(deskCaller(), "alpha", "pass:x", AnswerNo, "", &out)(s, at(10, 0)); err != nil {
			t.Fatal(err)
		}
		if out.Index != 1 || len(out.Older) != 1 || out.Older[0].Command != "old" || sess.Prompts[0].Answer != "" {
			t.Errorf("answered = %+v, prompts = %+v", out, sess.Prompts)
		}
	})
}

func TestStateGuard(t *testing.T) {
	s, _ := deskState()
	tests := []struct {
		name   string
		caller func() Caller
		want   string
	}{
		{"a desk", deskCaller, ""},
		{"not a desk", func() Caller { c := deskCaller(); c.SessionID = ""; return c }, "x: " + reasonNotDesk},
		{"not under the desk", func() Caller { c := deskCaller(); c.Chain = c.Chain[:1]; return c }, "x: " + reasonNotUnderDesk},
		{"no terminal", func() Caller { c := deskCaller(); c.Chain[1].TTY = 0; return c }, "x: " + reasonNoTerminal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Guard(tt.caller(), "x: ")
			switch {
			case tt.want == "" && err != nil:
				t.Fatal(err)
			case tt.want != "" && (err == nil || err.Error() != tt.want || !isRefusal(err)):
				t.Fatalf("error = %v, want refusal %q", err, tt.want)
			}
		})
	}
}

func TestAddRule(t *testing.T) {
	tests := []struct {
		name    string
		caller  func() Caller
		project string
		rule    string
		setup   func(rf *ReleaseFile)
		refused string
		bad     bool
	}{
		{name: "a new rule", rule: "Bash(go test ./...)"},
		{name: "a bare tool", rule: "WebFetch"},
		{name: "a duplicate", rule: "Bash(ls)", setup: func(rf *ReleaseFile) { rf.AllowRules = []string{"Bash(ls)"} }, refused: "already allows"},
		{name: "not a desk", rule: "Bash(ls)", refused: rulesPrefix + reasonNotDesk, caller: func() Caller { c := deskCaller(); c.SessionID = "o"; return c }},
		{name: "empty content", rule: "Bash()", bad: true},
		{name: "no closing paren", rule: "Bash(ls", bad: true},
		{name: "an unknown project", rule: "Bash(ls)", project: "zed", bad: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := deskState()
			if tt.setup != nil {
				tt.setup(rf)
			}
			caller, project := deskCaller, "alpha"
			if tt.caller != nil {
				caller = tt.caller
			}
			if tt.project != "" {
				project = tt.project
			}
			err := AddRule(caller(), project, tt.rule)(s, time.Time{})
			switch {
			case tt.refused != "":
				if err == nil || !isRefusal(err) || !strings.Contains(err.Error(), tt.refused) {
					t.Fatalf("error = %v, want refusal holding %q", err, tt.refused)
				}
			case tt.bad:
				if err == nil || isRefusal(err) {
					t.Fatalf("error = %v, want a bad request", err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Contains(rf.AllowRules, tt.rule) || !rf.Dirty {
					t.Errorf("allow rules = %q, dirty = %v", rf.AllowRules, rf.Dirty)
				}
			}
		})
	}
}

func TestRemoveRule(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		refused string
	}{
		{name: "a rule there", rule: "Bash(ls)"},
		{name: "a rule not there", rule: "Bash(pwd)", refused: "doesn't allow"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := deskState()
			rf.AllowRules = []string{"Bash(ls)", "Bash(make)"}
			err := RemoveRule(deskCaller(), "alpha", tt.rule)(s, time.Time{})
			if tt.refused != "" {
				if err == nil || !isRefusal(err) || !strings.Contains(err.Error(), tt.refused) {
					t.Fatalf("error = %v, want refusal holding %q", err, tt.refused)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(rf.AllowRules, []string{"Bash(make)"}) {
				t.Errorf("allow rules = %q", rf.AllowRules)
			}
		})
	}
}
