package service

import (
	"context"

	"svc/internal/domain/model"
)

func approve(request *model.ApproveWorkOrderRequest) error {
	return request.Validate() // want `GID-278: /domain/service calls ApproveWorkOrderRequest.Validate; transport requests and commands must be validated at ingress`
}

func complete(ctx context.Context, command model.CompleteWorkOrderCommand) error {
	return command.Validate(
		ctx,
	) // want `GID-278: /domain/service calls CompleteWorkOrderCommand.Validate; transport requests and commands must be validated at ingress`
}

func parse(request *model.ParsedRequest) (bool, error) {
	return request.Validate() // want `GID-278: /domain/service calls ParsedRequest.Validate; transport requests and commands must be validated at ingress`
}

func checkOrder(order *model.Order) error {
	return order.Validate()
}

func checkState(request *model.ListRequest) error {
	return request.ValidateState()
}

func ready(command *model.ReadyCommand) bool {
	return command.Validate()
}
