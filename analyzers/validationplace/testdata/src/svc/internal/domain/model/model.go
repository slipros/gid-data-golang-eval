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

// Validate may express an invariant on the model, so its declaration is legal;
// using it as the service input-validation boundary is not.
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
