// Негатив (граница): вложенные пакеты /domain/model/* — полноправный
// model-слой, объявление ошибок разрешено.
package filter

import "github.com/pkg/errors"

var ErrInvalidFilter = errors.New("invalid filter")

type Snapshots struct {
	IDs []string
}
