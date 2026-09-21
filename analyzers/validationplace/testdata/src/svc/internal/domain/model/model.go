package model

import "context"

type ApproveWorkOrderRequest struct{}

func (a *ApproveWorkOrderRequest) Validate() error { // want `GID-278:.*ApproveWorkOrderRequest`
	return nil
}

type CompleteWorkOrderCommand struct{}

func (CompleteWorkOrderCommand) Validate( // want `GID-278:.*CompleteWorkOrderCommand`
	context.Context,
) error {
	return nil
}

type ParsedRequest struct{}

func (*ParsedRequest) Validate() (bool, error) { // want `GID-278:.*ParsedRequest`
	return true, nil
}

type Order struct{}

// Validate is a domain-model invariant and is outside the suffix-based rule.
func (*Order) Validate() error {
	return nil
}

type ListRequest struct{}

// ValidateState is not the transport validator convention.
func (*ListRequest) ValidateState() error {
	return nil
}

type ReadyCommand struct{}

// Validate does not return an error and is outside the validation contract.
func (*ReadyCommand) Validate() bool {
	return true
}
