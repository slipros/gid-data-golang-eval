// Positive (GID-229): the client is isolated — domain is not available to it;
// for a third-party package (strconv) the rule does not apply.
package billing

import (
	"strconv"

	"svc/domain/model" // want `GID-229: package "svc/client/billing" must not import "svc/domain/model"\. Fix: the client has its own types; model <-> client DTO conversion lives at the consumer`
)

type Client struct{}

func (c *Client) Snapshot(id int) model.Snapshot {
	return model.Snapshot{ID: strconv.Itoa(id)}
}
