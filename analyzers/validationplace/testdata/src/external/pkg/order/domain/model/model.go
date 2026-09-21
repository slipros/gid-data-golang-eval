package model

type ExternalRequest struct{}

func (*ExternalRequest) Validate() error {
	return nil
}
