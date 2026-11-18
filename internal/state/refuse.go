package state

// RefusedError reports a request the state does not allow, as opposed to a
// broken file or a bad input. The command exits 2 for it.
type RefusedError struct {
	Msg string
}

func (e *RefusedError) Error() string { return e.Msg }

// Refuse returns a *RefusedError with the given message.
func Refuse(msg string) error { return &RefusedError{Msg: msg} }
