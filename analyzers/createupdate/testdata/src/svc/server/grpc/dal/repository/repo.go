// Boundary case: dal/repository nested under another layer (server/grpc) is
// NOT the repository layer — pathseg.HasLayer anchors the layer to the
// module root, so a Create*/Update* method here must NOT be flagged (would
// be a false positive under a plain path Contains).
package repository

import "context"

type Snapshot struct{ ID string }

type Job struct{}

func (j *Job) CreateJob(ctx context.Context, name string) (Snapshot, error) {
	return Snapshot{}, nil
}
