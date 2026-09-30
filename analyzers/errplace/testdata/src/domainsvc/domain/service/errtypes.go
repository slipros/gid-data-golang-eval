// Eval of GID-144 for error TYPES declared in /domain/service.
package service

import "domainsvc/domain/model"

// --- Positive: an error type outside model ---

type conflictError struct { // want `GID-144: error type "conflictError" is declared in "domainsvc/domain/service"\. Fix: declare the type in /domain/model, this layer only constructs it`
	id string
}

func (c *conflictError) Error() string { return "conflict " + c.id }

// --- Negative: constructing the model's error type ---

func invalid(field string) error {
	return model.ValidationError{Field: field}
}
