// Eval GID-124 (enum реализует String).
package enumstring

// --- Позитив: enum без String() ---

type SnapshotStatus string // want `GID-124: enum SnapshotStatus обязан реализовать метод String\(\) string`

const (
	SnapshotStatusPending   SnapshotStatus = "pending"
	SnapshotStatusUploading SnapshotStatus = "uploading"
)

// Граничный кейс: метод String с неправильной сигнатурой — не считается.
type JobStatus string // want `GID-124: enum JobStatus обязан реализовать метод String\(\) string`

const JobStatusActive JobStatus = "active"

func (j JobStatus) String(prefix string) string { return prefix + string(j) }

// --- Негатив: enum со String() ---

type UploadStatus string

const UploadStatusDone UploadStatus = "done"

func (u UploadStatus) String() string { return string(u) }

// --- Неприменимость ---

// string-тип без const-значений — не enum.
type Name string

// не-string тип с const — вне правила (enum у нас string-based).
type Level int

const LevelHigh Level = 1
