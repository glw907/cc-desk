package state

import "time"

// SameStart reports whether two process start times are the same moment.
// The boot time they come from is whole seconds, so two reads of it can
// differ by one; a gap under two seconds matches.
func SameStart(a, b time.Time) bool {
	d := a.Sub(b)
	return d > -2*time.Second && d < 2*time.Second
}
