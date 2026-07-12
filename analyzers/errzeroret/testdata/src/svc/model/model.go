// Stub of a domain/model value type for eval.
package model

// Result stands in for a value-type model result.
type Result struct {
	Value string
}

// TranscribeJobSource stands in for a string-based domain enum.
type TranscribeJobSource string

// TranscribeJobSource enum values — the Unspecified member is the zero value ("").
const (
	TranscribeJobSourceUnspecified TranscribeJobSource = ""
	TranscribeJobSourceUpload      TranscribeJobSource = "upload"
	TranscribeJobSourceRealtime    TranscribeJobSource = "realtime"
)

// Priority stands in for an int-based domain enum with a non-zero member.
type Priority int

// Priority enum values — Unspecified is zero (0), High is non-zero.
const (
	PriorityUnspecified Priority = 0
	PriorityHigh        Priority = 10
)
