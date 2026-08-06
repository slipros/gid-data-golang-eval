// Stub of google.golang.org/grpc/status for eval.
package status

import (
	stderrors "errors"

	"google.golang.org/grpc/codes"
)

// Error builds a status error from a code and a message.
func Error(c codes.Code, msg string) error {
	return stderrors.New(msg)
}

// Errorf builds a status error from a code and a formatted message.
func Errorf(c codes.Code, format string, args ...any) error {
	return stderrors.New(format)
}

// Status is a gRPC status — deliberately NOT an error (the real type exposes
// Err() error instead), which is what let a *status.Status mapper past the
// error-return-only discriminator of GID-242.
type Status struct {
	code codes.Code
	msg  string
}

// Code returns the status code.
func (s *Status) Code() codes.Code { return s.code }

// Message returns the status message.
func (s *Status) Message() string { return s.msg }

// Err converts the status into an error.
func (s *Status) Err() error { return stderrors.New(s.msg) }

// New builds a status from a code and a message.
func New(c codes.Code, msg string) *Status {
	return &Status{code: c, msg: msg}
}
