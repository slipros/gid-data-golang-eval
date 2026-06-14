// Eval GID-230: a service struct from settings.exclude is not checked.
package job

import (
	"svc/genproto/consentpb"

	"excluded/internal/server/grpc/job/handler"
)

// Job is listed in settings.exclude — non-Handler fields are not flagged.
type Job struct {
	consentpb.UnimplementedConsentServiceServer

	runner *handler.Run
}

func (j *Job) use() { _ = j.runner } //nolint:unused
