// Inapplicable: /domain/model/filter is a subpackage of model, not the layer root.
// pathseg.EndsWith(model) does not match here → GID-169 does not apply.
package filter

import "github.com/pkg/errors"

var ErrBadFilter = errors.New("bad filter")
