// Граница: *_test.go пропускается — объявление ошибки тут не нарушает GID-169.
package model

import "github.com/pkg/errors"

var errTestOnly = errors.New("test only")
