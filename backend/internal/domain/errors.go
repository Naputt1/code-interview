package domain

import "fmt"

// Error is a service error that carries the HTTP status and machine-readable
// code the transport should return. The service layer returns these; the
// transport maps anything else to a generic 500.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func NotFound(what string, id int64) *Error {
	return &Error{Status: 404, Code: "not_found", Message: fmt.Sprintf("%s %d not found", what, id)}
}

func Validation(message string) *Error {
	return &Error{Status: 400, Code: "validation_error", Message: message}
}

func Conflict(code, message string) *Error {
	return &Error{Status: 409, Code: code, Message: message}
}
