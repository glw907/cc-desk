package main

import (
	"strings"
	"testing"
)

func TestSlugOf(t *testing.T) {
	tests := []struct{ name, title, want string }{
		{"words", "Speed up the gate", "speed-up-the-gate"},
		{"runs of marks", "A -- B,  C!!", "a-b-c"},
		{"trimmed", "  --Edge--  ", "edge"},
		{"digits", "Pass 5 Live", "pass-5-live"},
		{"no letters", "!!! ???", ""},
		{"non-ASCII dropped", "café au lait", "caf-au-lait"},
		{"cut to 60", strings.Repeat("ab ", 40), strings.Repeat("ab-", 20)[:59]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slugOf(tt.title); got != tt.want {
				t.Errorf("slugOf(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}
