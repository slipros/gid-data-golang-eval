// Negative: /dal/entity is the home of dal errors, declaration is allowed here.
// Creation — via github.com/pkg/errors (GID-146).
package entity

import "github.com/pkg/errors"

var ErrRowNotFound = errors.New("row not found")

var ErrDuplicateKey = errors.New("duplicate key")
