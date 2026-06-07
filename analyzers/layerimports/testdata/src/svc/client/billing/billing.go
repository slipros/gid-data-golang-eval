// Позитив (GID-229): клиент изолирован — domain ему недоступен;
// сторонний пакет (strconv) — правило не применяется.
package billing

import (
	"strconv"

	"svc/domain/model" // want `GID-229: пакету "svc/client/billing" запрещён импорт "svc/domain/model" — у клиента свои типы: конвертация model <-> DTO клиента живёт у потребителя`
)

type Client struct{}

func (c *Client) Snapshot(id int) model.Snapshot {
	return model.Snapshot{ID: strconv.Itoa(id)}
}
