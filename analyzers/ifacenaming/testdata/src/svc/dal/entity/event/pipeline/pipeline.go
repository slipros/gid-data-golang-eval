// Eval GID-173 boundary: an "event" segment nested below another layer
// (dal/entity/event/pipeline) must NOT be classified as the /event/**
// layer. pathseg.Contains would match "event" anywhere in the path, wrongly
// putting this package in scope; the anchored pathseg.HasLayer requires
// "event" to be the leading segment after the module root, so this package
// is out of scope and the bare role name below is not flagged.
package pipeline

import "context"

type Service interface {
	Hello(ctx context.Context) error
}
