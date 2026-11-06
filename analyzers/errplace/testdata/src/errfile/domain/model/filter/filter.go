package filter

import "github.com/pkg/errors"

// Package error declarations use error.go even in nested packages.
var ErrBadFilter = errors.New("bad filter") // want `GID-169:`
