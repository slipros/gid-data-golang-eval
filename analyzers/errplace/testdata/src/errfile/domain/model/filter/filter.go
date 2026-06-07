// Неприменимость: /domain/model/filter — подпакет model, а не корень слоя.
// pathseg.EndsWith(model) тут не срабатывает → GID-169 не действует.
package filter

import "github.com/pkg/errors"

var ErrBadFilter = errors.New("bad filter")
