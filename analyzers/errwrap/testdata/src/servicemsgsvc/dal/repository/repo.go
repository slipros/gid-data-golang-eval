// Non-applicability of GID-237 outside /domain/service — errors.WithMessage in
// /dal/repository is unaffected by this rule.
package repository

import "github.com/pkg/errors"

type Repo struct{}

func (r *Repo) call() error { return nil }

func (r *Repo) f() error {
	err := r.call()
	return errors.WithMessage(err, "ctx")
}
