package state

import (
	"strings"
	"testing"
)

func TestCheckLoops(t *testing.T) {
	round := func(name, after string) string {
		a := ""
		if after != "" {
			a = `, "after": ["` + after + `"]`
		}
		return `{"name":"` + name + `","summary":"s","needs_geoff":true` + a + `}`
	}
	pass := func(name, rd, after string) string {
		a := ""
		if after != "" {
			a = `, "after": ["` + after + `"]`
		}
		return `{"name":"` + name + `","summary":"s","round":"` + rd + `","planned_minutes":10,"state":"not_started"` + a + `}`
	}
	build := func(rounds, passes []string) string {
		return `{"format":1,"project":"p","name":"n","repo":"r","rounds":[` + strings.Join(rounds, ",") +
			`],"passes":[` + strings.Join(passes, ",") + `]}`
	}
	tests := []struct {
		name string
		json string
		want string // empty means no error
	}{
		{"no loop", build([]string{round("a", ""), round("b", "a")}, []string{pass("x", "a", ""), pass("y", "b", "x")}), ""},
		{"pass loop", build([]string{round("a", "")}, []string{pass("x", "a", "y"), pass("y", "a", "x")}), "a loop in what rests on what"},
		{"self loop", build([]string{round("a", "")}, []string{pass("x", "a", "x")}), "pass x -> pass x"},
		{"round loop", build([]string{round("a", "b"), round("b", "a")}, nil), "a loop in what rests on what"},
		{"loop across rounds and passes", build([]string{round("a", ""), round("b", "a")}, []string{pass("pa", "a", "pb"), pass("pb", "b", "")}), "a loop in what rests on what"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRelease([]byte(tt.json))
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}
}

func TestCheckAnswers(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		want   string
	}{
		{"carried", `{"carried":{"words":"go","date":"2026-11-14","source":"chat"}}`, ""},
		{"both kinds", `{"ruling":"a","carried":{"words":"go","date":"d","source":"s"}}`, "one of the two"},
		{"neither", `{}`, "one of the two"},
		{"carried gap", `{"carried":{"words":"go"}}`, `needs "words", "date", and "source"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := `{"format":1,"project":"p","name":"n","repo":"r","questions":[{"id":"q","text":"t","options":["a","b"],"recommended":"a","answer":` + tt.answer + `}]}`
			_, err := DecodeRelease([]byte(j))
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}
}

func TestCheckRecommendedIsAnOption(t *testing.T) {
	j := `{"format":1,"project":"p","name":"n","repo":"r","questions":[{"id":"q","text":"t","options":["a"],"recommended":"z"}]}`
	if _, err := DecodeRelease([]byte(j)); err == nil || !strings.Contains(err.Error(), "not one of its options") {
		t.Errorf("error = %v", err)
	}
}
