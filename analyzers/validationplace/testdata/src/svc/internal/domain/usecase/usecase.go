package usecase

import "svc/internal/domain/model"

func approve(request *model.ApproveWorkOrderRequest) error {
	return request.Validate() // want `GID-278: /domain/usecase calls ApproveWorkOrderRequest.Validate; transport requests and commands must be validated at ingress`
}
