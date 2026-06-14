package llm

import (
	"errors"
	"fmt"
)

// Error carries the structured failure information surfaced by the
// LLM client. Callers can `errors.As` to inspect the kind.
type Error struct {
	Kind    ErrorKind
	Backend string // e.g. "anthropic", "openai" — empty when not applicable
	Status  int    // HTTP status when Kind == KindBadStatus
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("llm: %s [%s]: %s: %v", e.Kind, e.Backend, e.Message, e.Cause)
	}
	return fmt.Sprintf("llm: %s [%s]: %s", e.Kind, e.Backend, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

// ErrorKind discriminates the broad categories of llm.Error.
type ErrorKind string

const (
	KindUnknownProvider ErrorKind = "unknown_provider"
	KindNotConfigured   ErrorKind = "not_configured"
	KindTransport       ErrorKind = "transport"
	KindBadStatus       ErrorKind = "bad_status"
	KindDecode          ErrorKind = "decode"
	KindTimeout         ErrorKind = "timeout"
	KindInvalidRequest  ErrorKind = "invalid_request"
)

// IsKind reports whether err is an *Error of the given kind.
func IsKind(err error, kind ErrorKind) bool {
	var e *Error
	return errors.As(err, &e) && e.Kind == kind
}

// truncate returns body capped at 1 KiB for diagnostic messages.
func truncate(body string) string {
	const cap = 1024
	if len(body) <= cap {
		return body
	}
	return body[:cap] + "... [truncated]"
}
