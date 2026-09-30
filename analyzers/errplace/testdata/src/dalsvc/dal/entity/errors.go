// Negative: /dal/entity is the home of dal errors, declaration is allowed here.
// Creation — via github.com/pkg/errors (GID-146).
package entity

import "github.com/pkg/errors"

var ErrRowNotFound = errors.New("row not found")

var ErrDuplicateKey = errors.New("duplicate key")

// Negative: an error TYPE is allowed in entity — it is the layer's error home too.
type ClassifiedError struct {
	Sentinel error
	Cause    error
	Msg      string
}

func (c *ClassifiedError) Error() string { return c.Msg }

func (c *ClassifiedError) Unwrap() []error { return []error{c.Sentinel, c.Cause} }
