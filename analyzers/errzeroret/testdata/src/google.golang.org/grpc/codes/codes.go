// Stub of google.golang.org/grpc/codes for eval.
package codes

// Code is the gRPC status code type.
type Code uint32

const (
	OK Code = iota
	Canceled
	Unknown
	InvalidArgument
	DeadlineExceeded
	NotFound
	AlreadyExists
	PermissionDenied
	ResourceExhausted
	FailedPrecondition
	Aborted
	OutOfRange
	Unimplemented
	Internal
	Unavailable
	DataLoss
	Unauthenticated
)
