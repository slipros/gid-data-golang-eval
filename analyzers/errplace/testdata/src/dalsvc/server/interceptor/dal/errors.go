// Boundary: a package literally named "dal" nested under another layer
// (server/interceptor/dal) is NOT the /dal/entity layer — pathseg.HasLayer
// anchors the layer to the module root, unlike pathseg.Contains which would
// match "dal" anywhere in the path. Without the fix this package-level error
// and errors.New call would falsely trigger GID-145.
package dal

import "github.com/pkg/errors"

var ErrInterceptor = errors.New("interceptor failure")
