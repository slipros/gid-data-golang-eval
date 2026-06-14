// Eval of GID-176 (part 1): the /client boundary.
package client

import "github.com/pkg/errors"

type Client struct{}

func (c *Client) do() error { return nil }

// --- Positive: pass-through of an external call error ---

func (c *Client) badPassThrough() error {
	err := c.do()
	return err // want `GID-176: wrap with errors\.Wrap\. Fix: an error from the app boundary must collect stack and context`
}

// --- Negative: wrapped with Wrap ---

func (c *Client) goodWrap() error {
	err := c.do()
	return errors.Wrap(err, "http call")
}
