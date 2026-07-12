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
