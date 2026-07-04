// Positive: errors.go is no longer an allowed file by default —
// only error.go is; declaring errors here now violates GID-169.
package entity

import "github.com/pkg/errors"

var ErrRowNotFound = errors.New("row not found") // want `GID-169: error "ErrRowNotFound" is declared in errors\.go\. Fix: keep layer errors in error\.go`

var ErrDuplicateKey = errors.New("duplicate key") // want `GID-169: error "ErrDuplicateKey" is declared in errors\.go\. Fix: keep layer errors in error\.go`
