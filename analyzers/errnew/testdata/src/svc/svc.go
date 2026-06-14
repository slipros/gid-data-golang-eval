package svc

import (
	stderrors "errors"

	"github.com/pkg/errors"

	"svc/othernew"
)

// --- Negative: package-level var — the norm ---

// ErrNotFound — a static error in a single var. errors.New is legitimate here.
var ErrNotFound = errors.New("not found")

// A var block with several static errors — the norm.
var (
	ErrConflict = errors.New("conflict")
	ErrLocked   = errors.New("locked")
)

// --- Positive: errors.New in a func literal inside a package-level var ---

// makeErr — a package-level var holding a func literal; errors.New in its body
// is evaluated when the literal is called → runtime.
var makeErr = func() error {
	return errors.New("made at runtime") // want `GID-136: errors.New at runtime`
}

// --- Positive: errors.New in a function body ---

func loadSomething() error {
	return errors.New("load failed") // want `GID-136: errors.New at runtime`
}

// --- Boundary: errors.Errorf in a body — not GID-136 territory ---

func formatSomething(id int) error {
	return errors.Errorf("bad id %d", id) // dynamic context — legitimate (GID-144/145)
}

// --- Boundary: standard errors.New in a body — GID-146 territory, not GID-136 ---

func stdNew() error {
	return stderrors.New("std new") // std errors — not touched
}

// --- Boundary: a local New function from another package — not matched ---

func otherNew() error {
	return othernew.New("other") // not github.com/pkg/errors
}
