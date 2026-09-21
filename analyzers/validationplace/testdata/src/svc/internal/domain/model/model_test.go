package model

type fixtureRequest struct{}

func (*fixtureRequest) Validate() error {
	return nil
}
