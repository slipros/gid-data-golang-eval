// Негатив: /dal/entity — дом dal-ошибок, здесь объявление разрешено.
// Создание — через github.com/pkg/errors (GID-146).
package entity

import "github.com/pkg/errors"

var ErrRowNotFound = errors.New("row not found")

var ErrDuplicateKey = errors.New("duplicate key")
