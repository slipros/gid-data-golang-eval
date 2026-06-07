// dal-ошибки (статичные) — дом в /dal/entity.
package entity

import "github.com/pkg/errors"

var ErrNotFound = errors.New("not found")
