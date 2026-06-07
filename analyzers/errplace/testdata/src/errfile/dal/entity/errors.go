// Негатив: errors.go в /dal/entity — разрешённый файл.
package entity

import "github.com/pkg/errors"

var ErrRowNotFound = errors.New("row not found")

var ErrDuplicateKey = errors.New("duplicate key")
