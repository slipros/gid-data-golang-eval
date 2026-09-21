package model

type RefundRequest struct{}

func (*RefundRequest) Validate() error { // want `GID-278:.*RefundRequest`
	return nil
}
