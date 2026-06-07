// Позитив (GID-229): клиент изолирован — domain ему недоступен;
// сторонний пакет (strconv) — правило не применяется.
package billing

import (
	"strconv"

	"svc/domain/model" // want `GID-229: package "svc/client/billing" must not import "svc/domain/model"\. Fix: the client has its own types; model <-> client DTO conversion lives at the consumer`
)

type Client struct{}

func (c *Client) Snapshot(id int) model.Snapshot {
	return model.Snapshot{ID: strconv.Itoa(id)}
}
