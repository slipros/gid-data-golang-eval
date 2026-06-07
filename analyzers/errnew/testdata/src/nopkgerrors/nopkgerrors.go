// Package nopkgerrors не импортирует github.com/pkg/errors — правило неприменимо.
package nopkgerrors

import stderrors "errors"

// std errors.New в теле функции — зона GID-146, не GID-136.
func boom() error {
	return stderrors.New("boom")
}
