// Package entity — the CORE data layer: the module must not reach into it.
package entity

import "errors"

// Integration — the core record every integration type shares.
type Integration struct {
	ID string
}

// ErrNoResult — a sentinel of the core data layer: a module has its own.
var ErrNoResult = errors.New("no result")
