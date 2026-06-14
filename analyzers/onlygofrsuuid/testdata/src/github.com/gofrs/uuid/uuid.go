// A stub of the allowed library for the eval.
package uuid

type UUID [16]byte

func NewV7() (UUID, error) { return UUID{}, nil }

func Must(u UUID, err error) UUID { return u }
