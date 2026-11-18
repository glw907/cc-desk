package state

import (
	"fmt"
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
	for i, c := range r.Calls {
		for _, f := range []struct{ name, v string }{{"call", c.Call}, {"chosen", c.Chosen}, {"other", c.Other}, {"why", c.Why}} {
			if strings.TrimSpace(f.v) == "" {
				return nil, fmt.Errorf("calls[%d] is missing %q", i, f.name)
			}
		}
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

// reportable checks that some release of the project, live or done, holds
// the round, or exists when round is empty.
func (s *State) reportable(project, round string) error {
	known := false
	for _, r := range s.Releases {
		if r.Project != project {
			continue
		}
		known = true
		if round == "" || slices.ContainsFunc(r.Rounds, func(rd Round) bool { return rd.Name == round }) {
			return nil
		}
	}
	if known {
		return fmt.Errorf("round %q not found in project %q", round, project)
	}
	return fmt.Errorf("project %q not found", project)
}

// AddReport records the report file as the round's report, or as the
// release's when round is empty. A release that is done takes one as a live
// release does. It refuses a second report for the same project and round,
// and names a project or round that does not exist. When id is not nil it
// receives the new report's id.
func AddReport(project, round string, rf *ReportFile, id *string) Event {
	return func(s *State, now time.Time) error {
		if err := s.reportable(project, round); err != nil {
			return err
		}
		if slices.ContainsFunc(s.Reports.Reports, func(r Report) bool { return r.Project == project && r.Round == round }) {
			if round == "" {
				return Refuse(fmt.Sprintf("project %q already has a release report", project))
			}
			return Refuse(fmt.Sprintf("round %q of project %q already has a report", round, project))
		}
		rep := Report{
			ID: s.nextReportID(), Project: project, Round: round, Title: rf.Title, Summary: rf.Summary,
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
