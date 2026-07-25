// Non-applicability of GID-248 in a boundary layer: in /dal/repository an
// interface-method call is the external dependency (mechanism b of GID-176) —
// its error needs errors.Wrap, and GID-176 reports that; GID-248 stays silent
// so the two rules do not duplicate each other.
package repository

import "github.com/pkg/errors"

type Conn interface {
	Select(q string) error
}

type Repo struct {
	conn Conn
}

// buildQuery is a local pure helper — a same-module call, so its error is stacked.
func buildQuery() (string, error) { return "", nil }

func (r *Repo) boundaryCall() error {
	err := r.conn.Select("select 1")

	return errors.WithStack(err)
}

// A local helper of the same module is stacked even inside a boundary layer.
func (r *Repo) localHelper() error {
	_, err := buildQuery()

	return errors.WithStack(err) // want `GID-248: errors\.WithStack of an error that already carries a stack layers a second one\. Fix: return the error as is \(return err\)`
}
