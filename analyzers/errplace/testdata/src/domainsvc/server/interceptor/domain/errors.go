// Boundary: a package literally named "domain" nested under another layer
// (server/interceptor/domain) is NOT the /domain/model layer — pathseg.HasLayer
// anchors the layer to the module root, unlike pathseg.Contains which would
// match "domain" anywhere in the path. Without the fix this package-level
// error and errors.New call would falsely trigger GID-144.
package domain

import "github.com/pkg/errors"

var ErrInterceptor = errors.New("interceptor failure")
