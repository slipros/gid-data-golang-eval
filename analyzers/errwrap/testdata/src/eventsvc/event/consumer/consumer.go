// Eval of GID-176 (part 1, v2a): a direct external call (not through an
// injected interface) in /event/** is also a boundary — mechanism (a) applies
// in any layer, including one that already has its own interface-call
// boundary (mechanism b).
package consumer

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Event struct{}

type Consumer struct{}

// --- Positive: a direct external call (mechanism a) is a boundary here too ---

func (c *Consumer) badExternalCall(data []byte) error {
	var e Event
	err := json.Unmarshal(data, &e)
	return err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// --- Negative: wrapped with Wrap ---

func (c *Consumer) goodExternalWrap(data []byte) error {
	var e Event
	err := json.Unmarshal(data, &e)
	return errors.Wrap(err, "unmarshal event")
}

// --- Non-applicability: a local package function (not external, not interface) may use WithMessage ---

func buildKey(topic string) (string, error) { return topic, nil }

func (c *Consumer) goodLocalWithMessage() error {
	_, err := buildKey("topic")
	if err != nil {
		return errors.WithMessage(err, "build key")
	}
	return nil
}
