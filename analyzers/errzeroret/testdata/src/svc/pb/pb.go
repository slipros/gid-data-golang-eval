// Stub of a generated proto response type for eval.
package pb

// Resp stands in for a generated gRPC response message.
type Resp struct {
	Value string
}

// Status stands in for a generated proto3 enum (int32-based).
type Status int32

// Status enum values — the *_UNSPECIFIED member is the zero value (0).
const (
	Status_STATUS_UNSPECIFIED Status = 0
	Status_STATUS_ACTIVE      Status = 1
	Status_STATUS_CLOSED      Status = 2
)
