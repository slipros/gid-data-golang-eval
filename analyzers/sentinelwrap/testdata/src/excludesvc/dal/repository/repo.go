// Eval of GID-244 settings.exclude: an exempted method is not flagged even
// though it matches the anti-pattern.
package repository

import (
	"github.com/pkg/errors"

	"excludesvc/dal/entity"
)

type Conn interface {
	call() error
}

type Repo struct {
	conn Conn
}

func isNoResult(err error) bool { return err != nil }

// excludedMethod matches the anti-pattern but is listed in settings.exclude
// ("Repo.excludedMethod") — no diagnostic.
func (r *Repo) excludedMethod() error {
	err := r.conn.call()
	if err != nil {
		if isNoResult(err) {
			return errors.Wrap(entity.ErrNoResult, "op")
		}
		return errors.Wrap(err, "op")
	}
	return nil
}
