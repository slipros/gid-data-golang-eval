// Статичные ошибки model: package-level var и именованный error-тип.
// Объявления var НЕ задеваются GID-177 (это не return).
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")

// BigError — именованный error-тип.
type BigError struct {
	Code int
}

func (e BigError) Error() string { return "big error" }
