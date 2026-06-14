// Not applicable: the package is outside the repo/service/usecase/handler layers.
package util

type Payload struct{ Data string }

type Codec struct{}

func (c *Codec) Encode(in Payload) (*Payload, error) {
	return &in, nil
}
