// Негатив (GID-172) + неприменимость: client импортирует domain/model
// (ок) и сторонний пакет (правило не применяется).
package billing

import (
	"strconv"

	"svc/domain/model"
)

type Client struct{}

func (c *Client) Snapshot(id int) model.Snapshot {
	return model.Snapshot{ID: strconv.Itoa(id)}
}
