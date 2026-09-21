package service

import (
	"context"

	"svc/internal/domain/model"
	modelrequest "svc/internal/domain/model/request"
)

func approve(request *model.ApproveWorkOrderRequest) error {
	return request.Validate() // want `GID-278:.*ApproveWorkOrderRequest.Validate`
}

func complete(ctx context.Context, command model.CompleteWorkOrderCommand) error {
	return command.Validate(ctx) // want `GID-278:.*CompleteWorkOrderCommand.Validate`
}

func parse(request *model.ParsedRequest) (bool, error) {
	return request.Validate() // want `GID-278:.*ParsedRequest.Validate`
}

func checkOrder(order *model.Order) error {
	return order.Validate() // want `GID-278:.*Order.Validate`
}

func update(request *modelrequest.Update) error {
	return request.Validate() // want `GID-278:.*Update.Validate`
}

func checkState(request *model.ListRequest) error {
	return request.ValidateState()
}

func ready(command *model.ReadyCommand) bool {
	return command.Validate()
}
