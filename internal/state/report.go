package state

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// ReportFile is the file a report is recorded from. Its four keys are all
// required: a key that is missing or null leaves its field nil or empty,
// and a list of [] is accepted.
type ReportFile struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Calls   []Call `json:"calls"`
	Misses  []Miss `json:"misses"`
}

// checkCalls names the first call that has an empty part.
func checkCalls(calls []Call) error {
	for i, c := range calls {
		for _, f := range []struct{ name, v string }{{"call", c.Call}, {"chosen", c.Chosen}, {"other", c.Other}, {"why", c.Why}} {
			if strings.TrimSpace(f.v) == "" {
				return fmt.Errorf("calls[%d] is missing %q", i, f.name)
			}
		}
	}
	return nil
}

// checkRoundNotes refuses an empty context paragraph and an incomplete call.
func checkRoundNotes(context []string, calls []Call) error {
	for i, c := range context {
		if strings.TrimSpace(c) == "" {
			return fmt.Errorf("context[%d] is empty", i)
		}
	}
	return checkCalls(calls)
}

// DecodeReport strictly decodes a report file and names the first part
// that is missing or empty.
func DecodeReport(data []byte) (*ReportFile, error) {
	var r ReportFile
	if err := decodeStrict(data, &r); err != nil {
		return nil, err
	}
	for _, f := range []struct {
		name    string
		missing bool
	}{
		{"title", strings.TrimSpace(r.Title) == ""},
		{"summary", strings.TrimSpace(r.Summary) == ""},
		{"calls", r.Calls == nil},
		{"misses", r.Misses == nil},
	} {
		if f.missing {
			return nil, fmt.Errorf("missing required field %q", f.name)
		}
	}
	if err := checkCalls(r.Calls); err != nil {
		return nil, err
	}
	for i, m := range r.Misses {
		for _, f := range []struct{ name, v string }{{"miss", m.Miss}, {"cost", m.Cost}} {
			if strings.TrimSpace(f.v) == "" {
				return nil, fmt.Errorf("misses[%d] is missing %q", i, f.name)
			}
		}
	}
	return &r, nil
}

// validateReports checks what a loaded reports.json must hold: an id,
// project, title, summary, and sent time for each report, and no id twice.
// It also makes a missing list an empty one, so it prints as [].
func validateReports(rs *Reports) error {
	seen := map[string]bool{}
	for i := range rs.Reports {
		r := &rs.Reports[i]
		if r.ID == "" || r.Project == "" || r.Title == "" || r.Summary == "" || r.Sent.IsZero() {
			return fmt.Errorf("report %d is missing id, project, title, summary, or sent", i+1)
		}
		if seen[r.ID] {
			return fmt.Errorf("report id %q is used twice", r.ID)
		}
		seen[r.ID] = true
		if r.Calls == nil {
			r.Calls = []Call{}
		}
		if r.Misses == nil {
			r.Misses = []Miss{}
		}
	}
	return nil
}

func (s *State) nextReportID() string {
	for n := len(s.Reports.Reports) + 1; ; n++ {
		id := fmt.Sprintf("rep-%04d", n)
		if !slices.ContainsFunc(s.Reports.Reports, func(r Report) bool { return r.ID == id }) {
			return id
		}
	}
}

// ReportRelease gives the name of the release a report of the project
// belongs to: the live release, when it holds the round or the report is the
// release's, else the newest done release that holds the round. It names a
// project or round that does not exist.
func (s *State) ReportRelease(project, round string) (string, error) {
	known := false
	var newest *ReleaseFile
	for _, rf := range s.Releases {
		if rf.Project != project {
			continue
		}
		known = true
		if round != "" && !slices.ContainsFunc(rf.Rounds, func(rd Round) bool { return rd.Name == round }) {
			continue
		}
		if !rf.Done {
			return rf.Name, nil
		}
		if newest == nil || doneStamp(rf.Path) >= doneStamp(newest.Path) {
			newest = rf
		}
	}
	switch {
	case newest != nil:
		return newest.Name, nil
	case known && round != "":
		return "", fmt.Errorf("round %q not found in project %q", round, project)
	}
	return "", fmt.Errorf("project %q not found", project)
}

// doneStamp is the time stamp in a done release's file name.
func doneStamp(file string) string {
	base := strings.TrimSuffix(filepath.Base(file), ".json")
	return base[max(0, len(base)-len(doneStampLayout)):]
}

// AddReport records the report file as the round's report, or as the
// release's when round is empty, tagged with the release it belongs to
// (ReportRelease). A release that is done takes one as a live release does.
// It refuses a second report for the same release and round, and names a
// project or round that does not exist. When id is not nil it receives the
// new report's id.
func AddReport(project, round string, rf *ReportFile, id *string) Event {
	return func(s *State, now time.Time) error {
		release, err := s.ReportRelease(project, round)
		if err != nil {
			return err
		}
		if slices.ContainsFunc(s.Reports.Reports, func(r Report) bool { return r.Matches(project, release, round) }) {
			if round == "" {
				return Refuse(fmt.Sprintf("project %q already has a release report", project))
			}
			return Refuse(fmt.Sprintf("round %q of project %q already has a report", round, project))
		}
		rep := Report{
			ID: s.nextReportID(), Project: project, Release: release, Round: round, Title: rf.Title, Summary: rf.Summary,
			Calls: slices.Clone(rf.Calls), Misses: slices.Clone(rf.Misses), Sent: now.UTC(),
		}
		s.Reports.Reports = append(s.Reports.Reports, rep)
		s.reportsDirty = true
		if id != nil {
			*id = rep.ID
		}
		return nil
	}
}

// ReadReport marks a report read at now. A report already read keeps its
// first time and the event succeeds, so a second click changes nothing. It
// settles nothing, so a failed session list can't start a waiting pass.
func ReadReport(id string) Event {
	return func(s *State, now time.Time) error {
		s.noSettle = true
		i := slices.IndexFunc(s.Reports.Reports, func(r Report) bool { return r.ID == id })
		if i < 0 {
			return Refuse(fmt.Sprintf("report %q not found", id))
		}
		if s.Reports.Reports[i].Read != nil {
			return nil
		}
		t := now.UTC()
		s.Reports.Reports[i].Read = &t
		s.reportsDirty = true
		return nil
	}
}
