// Negative (GID-275): the package name is taken by another import, so every
// alias here is needed — aliased twice or once.
package service

import (
	"example.com/svc/internal/dal/repository/convert"
	serviceconvert "example.com/svc/internal/domain/service/convert"
)

var _ = convert.X + serviceconvert.X
