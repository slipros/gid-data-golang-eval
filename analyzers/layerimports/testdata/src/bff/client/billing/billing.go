// Eval for GID-267, non-applicability: the client of a BFF — it owns its
// types and knows nothing about the domain, exactly as GID-229 requires.
package billing

// Invoice — the client's own model, never a domain type.
type Invoice struct {
	ID string
}

type Client struct{}

func (c *Client) Invoice(id string) (Invoice, error) {
	return Invoice{ID: id}, nil
}
