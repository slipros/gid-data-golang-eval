// dal errors (static) — their home is /dal/entity.
package entity

import "github.com/pkg/errors"

var ErrNoResult = errors.New("no result")

// CabinetError — a typed dal error (a named error type is a "static" error for the rule).
type CabinetError struct{ Reason string }

func (c *CabinetError) Error() string { return c.Reason }
