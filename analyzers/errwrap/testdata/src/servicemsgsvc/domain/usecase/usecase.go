// Non-applicability of GID-237 outside /domain/service — errors.WithMessage in
// /domain/usecase is exactly where it belongs (service.md).
package usecase

import "github.com/pkg/errors"

type Usecase struct{}

func (u *Usecase) call() error { return nil }

func (u *Usecase) f() error {
	err := u.call()
	return errors.WithMessage(err, "ctx")
}
