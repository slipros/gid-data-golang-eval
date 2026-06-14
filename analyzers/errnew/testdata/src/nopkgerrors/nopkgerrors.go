// Package nopkgerrors does not import github.com/pkg/errors — the rule is inapplicable.
package nopkgerrors

import stderrors "errors"

// std errors.New in a function body — the domain of GID-146, not GID-136.
func boom() error {
	return stderrors.New("boom")
}
