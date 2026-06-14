// Inapplicable: /domain/service is NOT a model/entity layer root,
// GID-169 does not apply here (this is GID-144 territory). An error var in any
// service file must not produce a GID-169 diagnostic.
package service

import "github.com/pkg/errors"

var ErrServiceLocal = errors.New("service local")

type Snapshot struct{}
