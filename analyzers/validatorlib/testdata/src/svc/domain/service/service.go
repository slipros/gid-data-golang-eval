// Позитив: сторонний валидатор запрещён в любом пакете.
package service

import (
	validator "github.com/go-playground/validator/v10" // want `GID-164: сторонняя валидационная библиотека "github.com/go-playground/validator/v10" запрещена — используйте github\.com/raoptimus/validator\.go/v2`
)

type Snapshot struct {
	v *validator.Validate
}
