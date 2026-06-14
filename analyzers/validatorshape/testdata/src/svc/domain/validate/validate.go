// Eval GID-213: the shape of a validator in the validate layer.
package validate

import "context"

// jobReq — the request type for the validators.
type jobReq struct{ ID string }

// jobOpt — an extra parameter for the boundary case.
type jobOpt struct{}

// --- Positive class: violations ---

// No Validate method at all.
type CreateJob struct{} // want `GID-213: validator "CreateJob" must have a Validate\(ctx context.Context, req \*T\) error method\. Fix: add it`

// Validate without ctx as the first parameter.
type UpdateJob struct{} // want `GID-213: validator "UpdateJob" must have a Validate\(ctx context.Context, req \*T\) error method\. Fix: add it`

func (v *UpdateJob) Validate(req *jobReq) error { return nil }

// Validate returns (bool, error), not a single error.
type DeleteJob struct{} // want `GID-213: validator "DeleteJob" must have a Validate\(ctx context.Context, req \*T\) error method\. Fix: add it`

func (v *DeleteJob) Validate(ctx context.Context, req *jobReq) (bool, error) { return false, nil }

// --- Negative class: correct code ---

// A correct validator (pointer receiver) — ok.
type ListJobs struct{}

func (v *ListJobs) Validate(ctx context.Context, req *jobReq) error { return nil }

// A settings type (*Options) — not a validator, not flagged.
type ListJobsOptions struct{ Limit int }

// An unexported struct — not flagged.
type helper struct{} //nolint:unused

// --- Boundary class ---

// A value receiver — ok.
type GetJob struct{}

func (v GetJob) Validate(ctx context.Context, req *jobReq) error { return nil }

// Two parameters after ctx — ctx + a single error are sufficient, ok.
type PatchJob struct{}

func (v PatchJob) Validate(ctx context.Context, req *jobReq, opt jobOpt) error { return nil }
