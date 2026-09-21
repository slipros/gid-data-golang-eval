package service

import externalmodel "external/pkg/order/domain/model"

func checkExternal(request *externalmodel.ExternalRequest) error {
	return request.Validate()
}
