package state

import (
	"errors"
	"time"
)

// RecordRefusal appends a refusal to the one open session record whose
// session_id is sessionID, and sets found to whether it found one. It refuses
// a refusal with no tool or no time, since a record that holds one fails every
// later load. It skips settle, so a hook's commit starts no waiting pass and
// appends no forecast line.
func RecordRefusal(sessionID string, r Refusal, found *bool) Event {
	return func(s *State, now time.Time) error {
		if r.Tool == "" || r.Time.IsZero() {
			return errors.New("a refusal needs a tool and a time")
		}
		s.noSettle = true
		*found = false
		if sessionID == "" {
			return nil
		}
		r.Time = r.Time.UTC()
		s.LiveSessions(func(rf *ReleaseFile, record string, sess *Session) {
			if *found || sess.Ended != nil || sess.SessionID != sessionID {
				return
			}
			sess.Refusals = append(sess.Refusals, r)
			rf.Dirty = true
			*found = true
		})
		return nil
	}
}
