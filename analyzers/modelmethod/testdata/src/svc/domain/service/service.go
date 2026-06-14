// Eval for GID-195 (modelmethod): the service package is in the rule's scope.
package service

import (
	"strings"

	"svc/domain/model"
)

type SnapshotService struct {
	prefix string
}

// --- Positive: a private function with a single model parameter ---

func snapshotTitle(s *model.Snapshot) string { // want `GID-195: private function "snapshotTitle" works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return strings.ToUpper(s.Name)
}

// --- Positive: a model enum by value ---

func isDone(st model.Status) bool { // want `GID-195: private function "isDone" works only with the model.Status value\. Fix: this is model behaviour, make it a public method of that type`
	return st == model.StatusDone
}

// --- Positive: a method that does not use its receiver ---

func (s *SnapshotService) renderSnapshot(snap *model.Snapshot) string { // want `GID-195: method "renderSnapshot" ignores its receiver and works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return snap.ID + ":" + snap.Name
}

// --- Positive: an unnamed receiver ---

func (*SnapshotService) pingSnapshot(s *model.Snapshot) bool { // want `GID-195: method "pingSnapshot" ignores its receiver and works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return s.ID != ""
}

// --- Negative: the method uses its receiver — legitimately belongs to the struct ---

func (s *SnapshotService) decorate(snap *model.Snapshot) string {
	return s.prefix + snap.Name
}

// --- Negative: two values of the same type ---

func equalSnapshots(a, b *model.Snapshot) bool {
	return a.ID == b.ID
}

// --- Negative: the function depends on a package-level symbol of its own package ---

const serviceTag = "svc:"

func tagSnapshot(s *model.Snapshot) string {
	return serviceTag + s.Name
}

// --- Negative: the result is a type of its own package — not movable ---

func wrapSnapshot(s *model.Snapshot) *SnapshotService {
	return &SnapshotService{prefix: s.Name}
}

// --- Boundary: variadic ---

func joinSnapshots(ss ...model.Snapshot) string {
	names := make([]string, 0, len(ss))
	for i := range ss {
		names = append(names, ss[i].Name)
	}
	return strings.Join(names, ",")
}

// --- Boundary: a slice of a model type — not a single value ---

func firstName(ss []model.Snapshot) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[0].Name
}

// --- Boundary: a model-layer interface — cannot add a method ---

func validateAny(v model.Validator) bool {
	return v.Validate() == nil
}

// --- Boundary: a parameter of its own package's type — not model ---

type ServiceOptions struct {
	Name string
}

func optionsName(o *ServiceOptions) string {
	return o.Name
}

// --- Boundary: a generic function ---

func anyTitle[T any](v T) T {
	return v
}

// --- Boundary: an exported function — outside the rule ---

func TitleSnapshot(s *model.Snapshot) string {
	return s.Name
}
