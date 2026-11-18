package main

import (
	"os"
	"testing"
)

func TestParseHookPrompt(t *testing.T) {
	sample, err := os.ReadFile("../../e2e/testdata/prompts/hook-bash.json")
	if err != nil {
		t.Fatal(err)
	}
	real, err := os.ReadFile("../../e2e/testdata/prompts/hook-add-directories.json")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		in      string
		want    hookPrompt
		wantErr bool
	}{
		{name: "documented bash example", in: string(sample), want: hookPrompt{SessionID: "abc123", Tool: "Bash", Command: "rm -rf node_modules", Rule: "Bash(rm -rf node_modules)"}},
		{name: "real input with an addDirectories suggestion", in: string(real), want: hookPrompt{SessionID: "6f7ef48d-7f68-41a0-a8d7-ca8e9f8258a6", Tool: "Bash", Command: "date -u +%s > /tmp/cc-desk-fixture/.claude/jobs/b606693e/tmp/hookprobe/probe3.txt"}},
		{name: "file path", in: `{"session_id":"s","tool_name":"Edit","tool_input":{"file_path":"/a/b"}}`, want: hookPrompt{SessionID: "s", Tool: "Edit", Command: "/a/b"}},
		{name: "url", in: `{"session_id":"s","tool_name":"WebFetch","tool_input":{"url":"https://x.test"}}`, want: hookPrompt{SessionID: "s", Tool: "WebFetch", Command: "https://x.test"}},
		{name: "command wins over file path", in: `{"session_id":"s","tool_name":"T","tool_input":{"file_path":"f","command":"c"}}`, want: hookPrompt{SessionID: "s", Tool: "T", Command: "c"}},
		{name: "non-string values are skipped", in: `{"session_id":"s","tool_name":"T","tool_input":{"command":3,"file_path":"f"}}`, want: hookPrompt{SessionID: "s", Tool: "T", Command: "f"}},
		{name: "no tool input", in: `{"session_id":"s","tool_name":"T"}`, want: hookPrompt{SessionID: "s", Tool: "T"}},
		{name: "tool_input not an object", in: `{"session_id":"s","tool_name":"T","tool_input":"x"}`, wantErr: true},
		{name: "rule with no content", in: `{"session_id":"s","tool_name":"T","permission_suggestions":[{"type":"addRules","rules":[{"toolName":"Read"}]}]}`, want: hookPrompt{SessionID: "s", Tool: "T", Rule: "Read"}},
		{name: "first addRules entry", in: `{"session_id":"s","tool_name":"T","permission_suggestions":[{"type":"setMode"},{"type":"addRules","rules":[{"toolName":"A","ruleContent":"x"},{"toolName":"B"}]}]}`, want: hookPrompt{SessionID: "s", Tool: "T", Rule: "A(x)"}},
		{name: "not json", in: `nope`, wantErr: true},
		{name: "empty session id", in: `{"tool_name":"T"}`, wantErr: true},
		{name: "empty tool name", in: `{"session_id":"s"}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHookPrompt([]byte(tt.in))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, want error %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
