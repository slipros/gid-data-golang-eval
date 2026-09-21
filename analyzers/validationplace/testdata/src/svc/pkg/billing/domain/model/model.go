package model

type RefundRequest struct{}

func (*RefundRequest) Validate() error { // want `GID-278: Validate on domain model type RefundRequest owns transport validation`
	return nil
}
