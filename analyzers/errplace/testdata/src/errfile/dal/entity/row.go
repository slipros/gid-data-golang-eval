// Позитив: error-var в /dal/entity вне error.go/errors.go.
package entity

import "github.com/pkg/errors"

var ErrRowLocked = errors.New("row locked") // want `GID-169: ошибка "ErrRowLocked" объявлена в row\.go — ошибки слоя живут в error\.go`

type Row struct{ ID int64 }
