// Negative: errors.go in /dal/entity is an allowed file.
package entity

import "github.com/pkg/errors"

var ErrRowNotFound = errors.New("row not found")

var ErrDuplicateKey = errors.New("duplicate key")
