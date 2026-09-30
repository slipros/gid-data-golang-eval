// Eval of GID-145 for error TYPES declared in /dal/repository.
package repository

import "dalsvc/dal/entity"

// --- Positive: the incident — a struct with Error() on a pointer, Unwrap() []error ---

type classifiedError struct { // want `GID-145: error type "classifiedError" is declared in "dalsvc/dal/repository"\. Fix: declare the type in /dal/entity, this layer only constructs it`
	sentinel error
	cause    error
	msg      string
}

func (c *classifiedError) Error() string { return c.msg }

func (c *classifiedError) Unwrap() []error {
	if c.cause == nil {
		return []error{c.sentinel}
	}

	return []error{c.sentinel, c.cause}
}

// --- Positive: a value receiver, a non-struct underlying type, an exported name ---

type QueryError string // want `GID-145: error type "QueryError" is declared in`

func (q QueryError) Error() string { return string(q) }

// --- Positive: Error() promoted from an embedded error ---

type wrapped struct { // want `GID-145: error type "wrapped" is declared in`
	error
	table string
}

// --- Positive: a generic error type, in a grouped type block ---

type (
	genericError[T any] struct { // want `GID-145: error type "genericError" is declared in`
		value T
	}

	// Negative (inside the same block): a plain data type.
	rowKey struct{ id string }
)

func (g *genericError[T]) Error() string { return "generic" }

// --- Negative: building the entity type is exactly what the layer does ---

func newNotFound(msg string) error {
	return &entity.ClassifiedError{Sentinel: entity.ErrRowNotFound, Msg: msg}
}

// --- Boundary: Error with another signature is not an error type ---

type status struct{ code int }

func (s status) Error(code int) string { _ = s; return string(rune(code)) }

type counter struct{ n int }

func (c counter) Error() int { return c.n }

// --- Boundary: an interface contract and an alias declare no error value type ---

type notFounder interface {
	error
	NotFound() bool
}

type aliasedError = entity.ClassifiedError

// --- Boundary: a function-local type is not a package-level declaration ---

func local() error {
	type localErr struct{ error }

	return localErr{}
}
