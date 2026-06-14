// Eval GID-124 (an enum implements String).
package enumstring

// --- Positive: an enum without String() ---

type SnapshotStatus string // want `GID-124: enum SnapshotStatus must implement the String\(\) string method`

const (
	SnapshotStatusPending   SnapshotStatus = "pending"
	SnapshotStatusUploading SnapshotStatus = "uploading"
)

// Edge case: a String method with a wrong signature — does not count.
type JobStatus string // want `GID-124: enum JobStatus must implement the String\(\) string method`

const JobStatusActive JobStatus = "active"

func (j JobStatus) String(prefix string) string { return prefix + string(j) }

// --- Negative: an enum with String() ---

type UploadStatus string

const UploadStatusDone UploadStatus = "done"

func (u UploadStatus) String() string { return string(u) }

// --- Not applicable ---

// a string type without const values — not an enum.
type Name string

// a non-string type with const — outside the rule (our enums are string-based).
type Level int

const LevelHigh Level = 1
