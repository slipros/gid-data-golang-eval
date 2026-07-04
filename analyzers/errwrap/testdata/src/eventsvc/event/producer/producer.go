// Eval of GID-176 (part 1, v2b): the /event/** boundary. A Kafka producer talks
// to an external system through an injected interface — same shape as
// /client/** and /dal/repository.
package producer

import "github.com/pkg/errors"

// KafkaClient is the injected external dependency — calls to its methods are the boundary.
type KafkaClient interface {
	Send(topic string, msg []byte) error
}

type Producer struct {
	client KafkaClient
}

// --- Positive: pass-through of an error from an interface-method call at the /event boundary ---

func (p *Producer) badPassThrough() error {
	err := p.client.Send("topic", nil)
	return err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// --- Negative: wrapped with Wrap ---

func (p *Producer) goodWrap() error {
	err := p.client.Send("topic", nil)
	return errors.Wrap(err, "kafka send")
}
