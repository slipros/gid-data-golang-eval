// Eval of GID-177: static errors are wrapped with WithStack on return.
package service

import (
	"github.com/pkg/errors"

	"staticsvc/domain/model"
	"staticsvc/pkg/gderror"
)

type Service struct{}

// --- Positive: a package-level error var without a wrapper ---

func (s *Service) badVar() error {
	return model.ErrSnapshotNotFound // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Positive: the address of a named error type without a wrapper ---

func (s *Service) badPtrLit() error {
	return &model.BigError{Code: 1} // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Positive: a composite literal of a named error type without a wrapper ---

func (s *Service) badValueLit() error {
	return model.BigError{Code: 2} // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Positive: WithMessage attaches a message and no stack, so the static
// error goes out as stackless as a bare one — its only stack is the
// package-level var's (the package init, not the failure). The shape an HTTP
// client on the /client boundary shipped a whole package with
// (ad-cabinet-connector internal/client/yandexaudience, 2026-08-06). ---

func (s *Service) badWithMessage() error {
	return errors.WithMessage(model.ErrSnapshotNotFound, "ctx") // want `GID-177: a static error is returned without a stack — errors\.WithMessage attaches a message and no stack`
}

func (s *Service) badWithMessagef(code int) error {
	return errors.WithMessagef(model.ErrSnapshotNotFound, "status %d", code) // want `GID-177: a static error is returned without a stack — errors\.WithMessagef attaches a message and no stack`
}

// --- Boundary: stacked WithMessage calls still reach the static error underneath ---

func (s *Service) badNestedWithMessage() error {
	return errors.WithMessage(errors.WithMessage(model.ErrSnapshotNotFound, "inner"), "outer") // want `GID-177: a static error is returned without a stack — errors\.WithMessage attaches a message and no stack`
}

// --- Negative: WithMessage over a WRAPPED static error — the stack is already
// collected underneath, the message is just a message. ---

func (s *Service) goodWithMessageOverStack() error {
	return errors.WithMessage(errors.WithStack(model.ErrSnapshotNotFound), "ctx")
}

// --- Negative: WithMessage over an incoming non-static error — its stack was
// collected upstream; this is GID-176/GID-237 territory, not GID-177. ---

func (s *Service) goodWithMessageOverIncoming(err error) error {
	return errors.WithMessage(err, "ctx")
}

// --- Negative: WithStack / Wrap of a static error ---

func (s *Service) goodWithStack() error {
	return errors.WithStack(model.ErrSnapshotNotFound)
}

func (s *Service) goodWrap() error {
	return errors.Wrap(model.ErrSnapshotNotFound, "ctx")
}

// --- Boundary: returning an incoming non-static error — not GID-177 ---

func (s *Service) goodPassThrough(err error) error {
	return err
}

// --- Inapplicable: an excluded constructor collects the stack itself (settings.exclude) ---

func (s *Service) goodExcludedCtor() error {
	return gderror.NewUnhandledValueError("x")
}

// --- Inapplicable: a function without an error ---

func (s *Service) noError() int {
	return 0
}

// --- Positive: a func literal held by a package-level var is runtime code too ---
//
// The renderer table below is a var declaration, so its returns sit in no
// function DECLARATION at all. They are still returns, and the static error
// they hand back has no more stack than one returned from a method: the only
// stack it carries is the package init's, where the sentinel was built.

type renderer func(args []string) (string, error)

var renderers = map[string]renderer{
	"bad": func(args []string) (string, error) {
		if len(args) != 0 {
			return "", errors.WithMessage(model.ErrSnapshotNotFound, "takes no arguments") // want `GID-177: a static error is returned without a stack`
		}

		return "ok", nil
	},

	// --- Negative: the same literal wrapping the sentinel properly ---
	"good": func(args []string) (string, error) {
		if len(args) != 0 {
			return "", errors.Wrap(model.ErrSnapshotNotFound, "takes no arguments")
		}

		return "ok", nil
	},
}
