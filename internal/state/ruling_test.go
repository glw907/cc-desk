package state

import (
	"testing"
	"time"
)

func TestSameStart(t *testing.T) {
	base := time.Date(2026, 11, 18, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		b    time.Time
		want bool
	}{
		{"equal", base, true},
		{"1 second after", base.Add(time.Second), true},
		{"1 second before", base.Add(-time.Second), true},
		{"3 seconds after", base.Add(3 * time.Second), false},
		{"3 seconds before", base.Add(-3 * time.Second), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SameStart(base, tt.b); got != tt.want {
				t.Errorf("SameStart = %v, want %v", got, tt.want)
			}
		})
	}
}
