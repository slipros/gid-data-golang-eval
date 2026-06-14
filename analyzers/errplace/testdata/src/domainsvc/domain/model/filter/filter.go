// Negative (boundary): nested packages /domain/model/* are a full-fledged
// model layer, declaring errors is allowed.
package filter

import "github.com/pkg/errors"

var ErrInvalidFilter = errors.New("invalid filter")

type Snapshots struct {
	IDs []string
}
