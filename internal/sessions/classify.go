package sessions

import (
	"regexp"
	"slices"
	"strings"
)

// Label is the state the check reads from a row and its text. The strings
// are the labels of spikes/2-states/classify.py.
type Label string

const (
	LabelWorking          Label = "working"
	LabelIdle             Label = "idle"
	LabelLostConnection   Label = "lost connection"
	LabelPermissionPrompt Label = "permission prompt"
	LabelSignInExpired    Label = "expired sign-in"
	LabelUsageLimit       Label = "usage limit"
	LabelCantTell         Label = "can't tell"
)

// Job holds the two fields the check reads from a session's job state file.
type Job struct {
	Detail string
	Needs  string
}

var reasons = []struct {
	label   Label
	pattern *regexp.Regexp
}{
	{LabelSignInExpired, regexp.MustCompile(`Login expired`)},
	{LabelUsageLimit, regexp.MustCompile(`You've hit your .{0,30}limit`)},
	{LabelLostConnection, regexp.MustCompile(
		`Can't reach the API server|EAI_AGAIN|Couldn't connect through your proxy|ERR_PROXY_TUNNEL|proxy refused the tunnel`)},
}

// Classify labels a row. The job file is searched first, then the screen,
// each for the reasons in order, as classify.py does. A blocked row with no
// matching text is "can't tell".
func Classify(row Row, job Job, screen string) Label {
	if row.State == "working" {
		return LabelWorking
	}
	if row.State == "done" && row.Status == "idle" {
		return LabelIdle
	}
	if row.WaitingFor == "permission prompt" {
		return LabelPermissionPrompt
	}
	for _, text := range []string{job.Detail, job.Needs} {
		for _, r := range reasons {
			if r.pattern.MatchString(text) {
				return r.label
			}
		}
	}
	for _, r := range reasons {
		if r.pattern.MatchString(screen) {
			return r.label
		}
	}
	return LabelCantTell
}

var (
	csi   = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)
	osc   = regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)
	save  = regexp.MustCompile(`\x1b[78]`)
	space = regexp.MustCompile(`\s+`)
)

const lastLineRunes = 200

// LastLine gives the text that names a stop the check cannot read. It is the
// job file's detail when that is set, else the last non-empty line of the
// screen, with escape sequences removed. Runs of white space collapse to one
// space and the result keeps its last 200 runes. No text at all gives "no text".
func LastLine(detail, screen string) string {
	text := detail
	if strings.TrimSpace(text) == "" {
		text = ""
		stripped := save.ReplaceAllString(osc.ReplaceAllString(csi.ReplaceAllString(screen, ""), ""), "")
		lines := strings.Split(stripped, "\n")
		for _, l := range slices.Backward(lines) {
			if strings.TrimSpace(l) != "" {
				text = l
				break
			}
		}
	}
	text = strings.TrimSpace(space.ReplaceAllString(text, " "))
	if text == "" {
		return "no text"
	}
	if r := []rune(text); len(r) > lastLineRunes {
		text = string(r[len(r)-lastLineRunes:])
	}
	return text
}
