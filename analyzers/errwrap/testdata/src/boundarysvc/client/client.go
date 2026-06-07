// Eval GID-176 (часть 1): граница /client.
package client

import "github.com/pkg/errors"

type Client struct{}

func (c *Client) do() error { return nil }

// --- Позитив: pass-through ошибки внешнего вызова ---

func (c *Client) badPassThrough() error {
	err := c.do()
	return err // want `GID-176: wrap with errors\.Wrap\. Fix: an error from the app boundary must collect stack and context`
}

// --- Негатив: обёрнуто Wrap ---

func (c *Client) goodWrap() error {
	err := c.do()
	return errors.Wrap(err, "http call")
}
