// Package extlib — an external library (a path without service layer segments).
package extlib

// Encoder — an external library interface. Its use is allowed.
type Encoder interface {
	Encode(v any) error
}
