package state

import (
	"slices"
	"strings"
	"testing"
)

func TestAskRulesValidate(t *testing.T) {
	tests := []struct {
		name    string
		allow   []string
		ask     []string
		wantErr string
	}{
		{name: "none"},
		{name: "an ask rule", ask: []string{"Bash(date *)"}},
		{name: "distinct lists", allow: []string{"Bash(ls)"}, ask: []string{"Bash(date *)"}},
		{name: "an empty rule", ask: []string{""}, wantErr: "the rule is empty"},
		{name: "a rule in both", allow: []string{"Bash(ls)"}, ask: []string{"Bash(ls)"}, wantErr: "in both allow_rules and ask_rules"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, rf := deskState()
			rf.AllowRules, rf.AskRules = tt.allow, tt.ask
			err := validateRelease(&rf.Release)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatal(err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestAnswerPromptAskRule(t *testing.T) {
	s, rf := promptState()
	rf.AskRules = []string{"Bash(go test ./...)"}
	var out Answered
	if err := AnswerPrompt(deskCaller(), "alpha", "pass:x", AnswerYes, "Bash(go test ./...)", &out)(s, at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if got := rf.Passes[0].Session.Prompts[out.Index].Answer; got != AnswerYes {
		t.Errorf("answer = %q, want a one-time yes", got)
	}
	if len(rf.AllowRules) != 0 {
		t.Errorf("allow rules = %q, want none", rf.AllowRules)
	}
	if !slices.Equal(rf.AskRules, []string{"Bash(go test ./...)"}) {
		t.Errorf("ask rules = %q", rf.AskRules)
	}
}

func TestAddRuleRefusesAskRule(t *testing.T) {
	s, rf := deskState()
	rf.AskRules = []string{"Bash(ls)"}
	err := AddRule(deskCaller(), "alpha", "Bash(ls)")(s, at(10, 0))
	if !isRefusal(err) || !strings.Contains(err.Error(), "asks about") {
		t.Fatalf("error = %v, want a refusal that says it asks about the rule", err)
	}
	if len(rf.AllowRules) != 0 {
		t.Errorf("allow rules = %q", rf.AllowRules)
	}
}
