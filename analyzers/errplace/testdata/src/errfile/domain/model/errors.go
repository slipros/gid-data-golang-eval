// Негатив: errors.go — разрешённый файл, объявления ошибок здесь — ок.
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")

var ErrSnapshotExpired = errors.New("snapshot expired")
