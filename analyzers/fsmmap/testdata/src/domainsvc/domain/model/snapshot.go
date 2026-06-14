package model

type SnapshotStatus string

const (
	SnapshotStatusDraft    SnapshotStatus = "draft"
	SnapshotStatusActive   SnapshotStatus = "active"
	SnapshotStatusArchived SnapshotStatus = "archived"
)

type OtherStatus string

const OtherStatusUnknown OtherStatus = "unknown"

// Positive: exported map[E][]E over a local string enum.
var SnapshotStatusTransitions = map[SnapshotStatus][]SnapshotStatus{ // want `GID-231: FSM transition map SnapshotStatusTransitions is exported\. Fix: make it unexported: var snapshotStatusTransitions = map\[Status\]\[\]Status\{\.\.\.\}`
	SnapshotStatusDraft:  {SnapshotStatusActive, SnapshotStatusArchived},
	SnapshotStatusActive: {SnapshotStatusArchived},
}

// Positive: exported map[E]map[E]struct{} (set form).
var StatusTransitionSet = map[SnapshotStatus]map[SnapshotStatus]struct{}{ // want `GID-231: FSM transition map StatusTransitionSet is exported\. Fix: make it unexported: var statusTransitionSet = map\[Status\]\[\]Status\{\.\.\.\}`
	SnapshotStatusDraft: {SnapshotStatusActive: {}},
}

// Positive: exported map[E]map[E]bool (flag form).
var StatusTransitionFlags = map[SnapshotStatus]map[SnapshotStatus]bool{ // want `GID-231: FSM transition map StatusTransitionFlags is exported\. Fix: make it unexported: var statusTransitionFlags = map\[Status\]\[\]Status\{\.\.\.\}`
	SnapshotStatusDraft: {SnapshotStatusActive: true},
}

type transitionMap map[SnapshotStatus][]SnapshotStatus

// Positive: exported var of a named map type whose underlying is a transition map.
var NamedTransitions = transitionMap{ // want `GID-231: FSM transition map NamedTransitions is exported\. Fix: make it unexported: var namedTransitions = map\[Status\]\[\]Status\{\.\.\.\}`
	SnapshotStatusDraft: {SnapshotStatusActive},
}

// Negative: unexported transition map — the canonical styleguide form.
var snapshotStatusTransitions = map[SnapshotStatus][]SnapshotStatus{
	SnapshotStatusDraft:  {SnapshotStatusActive, SnapshotStatusArchived},
	SnapshotStatusActive: {SnapshotStatusArchived},
}

// CanTransitionTo is the canonical consumer of the unexported map.
func (s SnapshotStatus) CanTransitionTo(target SnapshotStatus) bool {
	allowed, ok := snapshotStatusTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}
