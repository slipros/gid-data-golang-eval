// domain-ошибки (статичные) — дом в /domain/model.
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")
