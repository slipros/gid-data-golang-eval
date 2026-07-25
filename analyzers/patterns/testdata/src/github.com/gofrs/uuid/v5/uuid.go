// Minimal github.com/gofrs/uuid stub for eval.
package uuid

type UUID [16]byte

func (u UUID) IsNil() bool { return u == UUID{} }

func NewV1() (UUID, error)            { return UUID{}, nil }
func NewV3(ns UUID, name string) UUID { return UUID{} }
func NewV4() (UUID, error)            { return UUID{}, nil }
func NewV5(ns UUID, name string) UUID { return UUID{} }
func NewV6() (UUID, error)            { return UUID{}, nil }
func NewV7() (UUID, error)            { return UUID{}, nil }

func Must(u UUID, err error) UUID { return u }
