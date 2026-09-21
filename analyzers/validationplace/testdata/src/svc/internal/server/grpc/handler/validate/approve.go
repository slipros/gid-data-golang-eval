package validate

import "context"

type ApproveWorkOrderRequest struct{}

type ApproveWorkOrder struct{}

func (*ApproveWorkOrder) Validate(context.Context, *ApproveWorkOrderRequest) error {
	return nil
}
