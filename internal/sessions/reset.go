package sessions

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// resetPattern reads the reset clause of a limit sentence: an optional
// "<Mon> <D>, " date, an hour with optional minutes, am or pm, and a zone in
// parentheses.
var resetPattern = regexp.MustCompile(`resets (?:([A-Z][a-z]{2}) (\d{1,2}), )?(\d{1,2})(?::(\d{2}))?\s?([AaPp][Mm]) \(([^)]+)\)`)

// ParseReset reads the time a usage stop says its limit resets. A dated
// reset takes the year of began, the moment the stop opened, in the named
// zone, and the next year when that date falls more than 180 days before
// began. A time alone is the first such time after began. The second result
// is false when the text has no reset clause, the zone is unknown, or a
// field is out of range.
func ParseReset(text string, began time.Time) (time.Time, bool) {
	m := resetPattern.FindStringSubmatch(text)
	if m == nil {
		return time.Time{}, false
	}
	loc, err := time.LoadLocation(m[6])
	if err != nil {
		return time.Time{}, false
	}
	hour, _ := strconv.Atoi(m[3])
	if hour < 1 || hour > 12 {
		return time.Time{}, false
	}
	minute := 0
	if m[4] != "" {
		if minute, _ = strconv.Atoi(m[4]); minute > 59 {
			return time.Time{}, false
		}
	}
	hour %= 12
	if strings.EqualFold(m[5], "pm") {
		hour += 12
	}

	local := began.In(loc)
	if m[1] == "" {
		at := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc)
		if !at.After(began) {
			at = time.Date(local.Year(), local.Month(), local.Day()+1, hour, minute, 0, 0, loc)
		}
		return at, true
	}
	month, err := time.Parse("Jan", m[1])
	if err != nil {
		return time.Time{}, false
	}
	day, _ := strconv.Atoi(m[2])
	at := time.Date(local.Year(), month.Month(), day, hour, minute, 0, 0, loc)
	if at.Month() != month.Month() {
		return time.Time{}, false
	}
	if began.Sub(at) > 180*24*time.Hour {
		at = time.Date(local.Year()+1, month.Month(), day, hour, minute, 0, 0, loc)
	}
	return at, true
}
