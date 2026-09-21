package handler

import "svc/internal/domain/model"

func validate(request *model.ApproveWorkOrderRequest) error {
	return request.Validate()
}
