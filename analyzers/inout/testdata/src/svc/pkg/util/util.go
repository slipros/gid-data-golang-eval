// Неприменимость: пакет вне слоёв repo/service/usecase/handler.
package util

type Payload struct{ Data string }

type Codec struct{}

func (c *Codec) Encode(in Payload) (*Payload, error) {
	return &in, nil
}
