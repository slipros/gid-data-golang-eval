// Eval of GID-176 (part 1): the /client boundary.
// The boundary is an interface-method call (an injected transport dependency).
package client

import "github.com/pkg/errors"

// Transport is the injected external dependency — calls to its methods are the boundary.
type Transport interface {
	do() error
}

type Client struct {
	transport Transport
}

// --- Positive: pass-through of an external call error ---

func (c *Client) badPassThrough() error {
	err := c.transport.do()
	return err // want `GID-176: an error from the app boundary must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// --- Negative: wrapped with Wrap ---

func (c *Client) goodWrap() error {
	err := c.transport.do()
	return errors.Wrap(err, "http call")
}
