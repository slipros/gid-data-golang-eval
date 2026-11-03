package svc

import pkgerrors "github.com/pkg/errors"

func pkgErrorsWithoutGuard(c client) error {
	err := c.failure()
	if pkgerrors.Is(err, ErrRate) { // want `GID-279: the error "err" is classified before checking it for nil`
		return ErrRejected
	}
	return err
}

func pkgErrorsWithGuard(c client) error {
	err := c.failure()
	if err != nil {
		var target *rateError
		if pkgerrors.As(err, &target) {
			return target
		}
		return err
	}
	return nil
}
