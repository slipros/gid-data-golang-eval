// Негатив: /domain/model — дом domain-ошибок, здесь объявление разрешено.
// Создание — через github.com/pkg/errors (GID-146).
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")

var ErrSnapshotExpired = errors.New("snapshot expired")
