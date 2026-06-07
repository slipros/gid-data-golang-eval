// Позитив: сторонний валидатор запрещён в любом пакете.
package service

import (
	validator "github.com/go-playground/validator/v10" // want `GID-164: third-party validation library "github.com/go-playground/validator/v10" is forbidden\. Fix: use github\.com/raoptimus/validator\.go/v2`
)

type Snapshot struct {
	v *validator.Validate
}
