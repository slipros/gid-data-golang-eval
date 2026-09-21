package service

import "svc/pkg/billing/domain/model"

func refund(request *model.RefundRequest) error {
	return request.Validate() // want `GID-278: /domain/service calls RefundRequest.Validate; transport requests and commands must be validated at ingress`
}
