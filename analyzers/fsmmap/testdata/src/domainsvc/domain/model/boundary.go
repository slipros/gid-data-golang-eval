package model

// Boundary: value is not the enum — a lookup table, not a transition map.
var StatusLabels = map[SnapshotStatus][]string{
	SnapshotStatusDraft: {"draft"},
}

// Boundary: key and value are two different enums — not a transition map.
var MixedTransitions = map[SnapshotStatus][]OtherStatus{
	SnapshotStatusDraft: {OtherStatusUnknown},
}

type Priority int

const (
	PriorityLow  Priority = 1
	PriorityHigh Priority = 2
)

// Boundary: int-based type with consts is not a string enum.
var PriorityTransitions = map[Priority][]Priority{
	PriorityLow: {PriorityHigh},
}

type RawTag string

// Boundary: named string type without const values is not an enum.
var RawTransitions = map[RawTag][]RawTag{
	"a": {"b"},
}

// Boundary: inner map value is neither struct{} nor bool.
var StatusMatrix = map[SnapshotStatus]map[SnapshotStatus]int{
	SnapshotStatusDraft: {SnapshotStatusActive: 1},
}
