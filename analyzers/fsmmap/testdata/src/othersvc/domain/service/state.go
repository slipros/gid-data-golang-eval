package service

type JobStatus string

const (
	JobStatusNew  JobStatus = "new"
	JobStatusDone JobStatus = "done"
)

// Non-applicability: the same exported transition map outside /domain/model
// is not flagged by GID-231 (where FSM logic belongs is another rule's zone).
var JobStatusTransitions = map[JobStatus][]JobStatus{
	JobStatusNew: {JobStatusDone},
}
