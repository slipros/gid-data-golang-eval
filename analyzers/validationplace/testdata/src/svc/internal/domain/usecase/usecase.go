package usecase

import "svc/internal/domain/model"

func approve(request *model.ApproveWorkOrderRequest) error {
	return request.Validate() // want `GID-278:.*ApproveWorkOrderRequest.Validate`
}
