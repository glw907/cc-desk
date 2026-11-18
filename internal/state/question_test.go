package state

import (
	"strings"
	"testing"
)

const goodQuestion = `{"id":"q9","text":"Which?","options":["a","b"],"recommended":"a",
"background":["One."],"causes":{"a":"A causes this.","b":"B causes that."},"reason":"Because.","passes":["p1"]}`

func TestDecodeQuestion(t *testing.T) {
	tests := []struct {
		name, json, want string
	}{
		{"good", goodQuestion, ""},
		{"no background", strings.Replace(goodQuestion, `"background":["One."],`, "", 1), `"background"`},
		{"null causes", strings.Replace(goodQuestion, `"causes":{"a":"A causes this.","b":"B causes that."}`, `"causes":null`, 1), `"causes"`},
		{"no reason", strings.Replace(goodQuestion, `"reason":"Because.",`, "", 1), `"reason"`},
		{"no passes", strings.Replace(goodQuestion, `"passes":["p1"]`, `"passes":[]`, 1), `"passes"`},
		{"blank paragraph", strings.Replace(goodQuestion, `["One."]`, `["One."," "]`, 1), "paragraph 2"},
		{"unknown key", strings.Replace(goodQuestion, `"id":"q9"`, `"id":"q9","extra":1`, 1), "extra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeQuestion([]byte(tt.json))
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one naming %s", err, tt.want)
			}
		})
	}
}

func TestAddQuestion(t *testing.T) {
	qf, err := DecodeQuestion([]byte(goodQuestion))
	if err != nil {
		t.Fatal(err)
	}
	s := loaded(t)
	if err := AddQuestion("alpha", "r1", qf)(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	rf := s.Releases[0]
	if q := rf.question("q9"); q == nil || q.Reason != "Because." || len(q.Causes) != 2 {
		t.Fatalf("question after add = %+v", q)
	}
	if got := rf.Passes[0].Questions; len(got) != 1 || got[0] != "q9" || !rf.Dirty {
		t.Errorf("pass questions = %v, dirty %v", got, rf.Dirty)
	}
	if err := AddQuestion("alpha", "r1", qf)(s, at(9, 0)); !isRefusal(err) {
		t.Errorf("a second add: error = %v, want a refusal", err)
	}

	tests := []struct {
		name, project, round, pass string
	}{
		{"no project", "nope", "r1", "p1"},
		{"no round", "alpha", "r9", "p1"},
		{"no pass", "alpha", "r1", "p9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := *qf
			c.Passes = []string{tt.pass}
			err := AddQuestion(tt.project, tt.round, &c)(loaded(t), at(9, 0))
			if err == nil || isRefusal(err) {
				t.Errorf("error = %v, want a plain error", err)
			}
		})
	}

	t.Run("started pass", func(t *testing.T) {
		s := loaded(t)
		s.Releases[0].Passes[0].State = Building
		if err := AddQuestion("alpha", "r1", qf)(s, at(9, 0)); !isRefusal(err) {
			t.Errorf("error = %v, want a refusal", err)
		}
	})
}

func TestValidateCauses(t *testing.T) {
	tests := []struct {
		name   string
		causes map[string]string
		ok     bool
	}{
		{"none", nil, true},
		{"all options", map[string]string{"a": "x", "b": "y"}, true},
		{"a stranger", map[string]string{"a": "x", "b": "y", "c": "z"}, false},
		{"one missing", map[string]string{"a": "x"}, false},
		{"one empty", map[string]string{"a": "x", "b": ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCauses(&Question{ID: "q", Options: []string{"a", "b"}, Causes: tt.causes})
			if (err == nil) != tt.ok {
				t.Errorf("error = %v, ok want %v", err, tt.ok)
			}
		})
	}
}

func TestStartRefusesGuessOnNeverGuess(t *testing.T) {
	s := loaded(t)
	rf := s.Releases[0]
	rf.Questions = append(rf.Questions, Question{ID: "q1", Text: "T?", Options: []string{"a", "b"}, Recommended: "a", NeverGuess: true})
	rf.Passes[0].Questions = []string{"q1"}
	err := StartOnGuess("alpha", "p1", "q1")(s, at(9, 0))
	if !isRefusal(err) || !strings.Contains(err.Error(), "q1") {
		t.Errorf("error = %v, want a refusal naming q1", err)
	}
}
