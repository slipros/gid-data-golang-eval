// Eval for GID-274: an enum/discriminator switch must return the exact
// wrapped unhandled-value error from its default clause.
package svc

import (
	gderror "gderror"

	"github.com/pkg/errors"
)

// Status is a named string enum.
type Status string

const (
	StatusDraft Status = "draft"
	StatusReady Status = "ready"
)

// EventKind is a named discriminator used by Event.
type EventKind string

const (
	EventKindCreated EventKind = "created"
	EventKindDeleted EventKind = "deleted"
)

type Event struct {
	Kind EventKind
}

// --- Class 1: positive ---

func statusLabel(status Status) error {
	switch status { // want `GID-274: enum/discriminator switch default must directly return`
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	}

	return nil
}

func eventLabel(event Event) error {
	switch event.Kind { // want `GID-274: enum/discriminator switch default must directly return`
	case EventKindCreated:
		return nil
	case EventKindDeleted:
		return nil
	}

	return nil
}

func statusLabelWithFallback(status Status) error {
	switch status { // want `GID-274: enum/discriminator switch default must directly return`
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	default:
		return nil
	}
}

func statusLabelWithBareUnhandledError(status Status) error {
	switch status { // want `GID-274: enum/discriminator switch default must directly return`
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	default:
		return gderror.NewUnhandledValueError(status)
	}
}

func statusLabelWithDifferentValue(status Status) error {
	other := StatusDraft
	switch status { // want `GID-274: enum/discriminator switch default must directly return`
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	default:
		return errors.WithStack(gderror.NewUnhandledValueError(other))
	}
}

func statusLabelWithAnotherHandler(status Status) error {
	switch status { // want `GID-274: enum/discriminator switch default must directly return`
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	default:
		return errors.New("unknown status")
	}
}

func statusLabelWithExtraStatement(status Status) error {
	switch status { // want `GID-274: enum/discriminator switch default must directly return`
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	default:
		_ = status
		return errors.WithStack(gderror.NewUnhandledValueError(status))
	}
}

// --- Class 2: negative ---

func statusLabelWithUnhandledError(status Status) error {
	switch status {
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	default:
		return errors.WithStack(gderror.NewUnhandledValueError(status))
	}
}

func eventLabelWithUnhandledError(event Event) error {
	switch event.Kind {
	case EventKindCreated:
		return nil
	case EventKindDeleted:
		return nil
	default:
		return errors.WithStack(gderror.NewUnhandledValueError(event.Kind))
	}
}

// A call through a function value is not the direct static call required by
// the contract.
func statusLabelWithIndirectCall(status Status) error {
	withStack := errors.WithStack
	switch status { // want `GID-274: enum/discriminator switch default must directly return`
	case StatusDraft:
		return nil
	case StatusReady:
		return nil
	default:
		return withStack(gderror.NewUnhandledValueError(status))
	}
}

// --- Class 3: boundary ---

type Name string

type NumericKind int

const (
	NumericKindOne NumericKind = 1
	NumericKindTwo NumericKind = 2
)

func plainStringLabel(value string) string {
	switch value {
	case "draft":
		return "draft"
	}

	return ""
}

func plainIntLabel(value int) string {
	switch value {
	case 1:
		return "one"
	}

	return ""
}

func namedIntLabel(value NumericKind) string {
	switch value {
	case NumericKindOne:
		return "one"
	case NumericKindTwo:
		return "two"
	}

	return ""
}

func namedStringWithoutConstants(value Name) string {
	switch value {
	case "draft":
		return "draft"
	}

	return ""
}

func conditionSwitch(value bool) string {
	switch {
	case value:
		return "true"
	}

	return "false"
}
