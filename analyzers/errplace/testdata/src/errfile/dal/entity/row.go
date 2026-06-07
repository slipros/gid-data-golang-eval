// Позитив: error-var в /dal/entity вне error.go/errors.go.
package entity

import "github.com/pkg/errors"

var ErrRowLocked = errors.New("row locked") // want `GID-169: error "ErrRowLocked" is declared in row\.go\. Fix: keep layer errors in error\.go`

type Row struct{ ID int64 }
