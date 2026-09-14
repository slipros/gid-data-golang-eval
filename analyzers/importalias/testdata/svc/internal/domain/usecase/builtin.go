// Negative (GID-275): a package named like a predeclared identifier — without
// the alias it would hide the built-in error type from the whole file, even
// where the import is used at package level, outside any function.
package usecase

import (
	modelerror "example.com/svc/internal/domain/model/error"
)

var code = modelerror.X

func Fail() error {
	_ = code

	return nil
}
