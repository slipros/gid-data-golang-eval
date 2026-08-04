// Eval of GID-255 negative: a real client — a type whose method calls the
// external API and returns the client's own model. The constructor is fine
// here precisely because it is not the only thing in the package.
package payments

import "context"

// Payment is the client's own model — the consumer never sees the transport.
type Payment struct {
	ID     string
	Amount int64
}

// Client talks to the payments API.
type Client struct {
	addr string
}

// NewClient creates the payments client.
func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

// Payment fetches one payment.
func (c *Client) Payment(ctx context.Context, id string) (Payment, error) {
	_ = ctx
	return Payment{ID: id}, nil
}
