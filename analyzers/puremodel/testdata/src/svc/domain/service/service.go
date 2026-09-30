// Eval for GID-278 (puremodel): the service package is in the rule's scope.
package service

import (
	"context"
	"strings"

	"svc/domain/model"
)

type SegmentChange struct{}

type Notifier struct {
	prefix string
}

// --- Positive: the incident — an exported method over two model values, receiver unused ---

func (s *SegmentChange) SegmentChanged(current, snapshot *model.SegmentMetric) bool { // want `GID-278: method "SegmentChanged" uses neither its receiver nor the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric and drop the service that only carries it`
	if snapshot.IsEmpty() {
		return true
	}
	if current.RowCount != snapshot.RowCount {
		return true
	}

	return current.Checksum != snapshot.Checksum
}

// --- Positive: an unnamed receiver ---

func (*SegmentChange) Same(a, b *model.SegmentMetric) bool { // want `GID-278: method "Same" uses neither its receiver nor the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric and drop the service that only carries it`
	return a.Checksum == b.Checksum
}

// --- Positive: an exported method over a single model value ---

func (n *Notifier) Title(m *model.SegmentMetric) string { // want `GID-278: method "Title" uses neither its receiver nor the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric and drop the service that only carries it`
	return strings.ToUpper(m.Checksum)
}

// --- Positive: a private method over two model values (GID-195 judges only one) ---

func (n *Notifier) grew(before, after model.SegmentMetric) bool { // want `GID-278: method "grew" uses neither its receiver nor the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric and drop the service that only carries it`
	return after.RowCount > before.RowCount
}

// --- Positive: a model value plus plain data; an enum is a model type too ---

func (n *Notifier) Reached(st model.Status, limit int) bool { // want `GID-278: method "Reached" uses neither its receiver nor the package, it is pure logic over model.Status\. Fix: make it a public method of model.Status and drop the service that only carries it`
	return st == model.StatusDone && limit > 0
}

// --- Positive: free functions, exported and private with two parameters ---

func SegmentDiffers(a, b *model.SegmentMetric) bool { // want `GID-278: function "SegmentDiffers" uses nothing of the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric`
	return a.Checksum != b.Checksum
}

func rowsDiffer(a, b model.SegmentMetric) bool { // want `GID-278: function "rowsDiffer" uses nothing of the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric`
	return a.RowCount != b.RowCount
}

// --- Negative: the method uses its receiver ---

func (n *Notifier) Decorate(m *model.SegmentMetric) string {
	return n.prefix + m.Checksum
}

// --- Negative: a private function with one model parameter — GID-195's, not reported twice ---

func onlyChecksum(m *model.SegmentMetric) string {
	return m.Checksum
}

// --- Negative: the function depends on a package-level symbol ---

const serviceTag = "svc:"

func Tag(m *model.SegmentMetric) string {
	return serviceTag + m.Checksum
}

// --- Negative: a package type in the result — moving it would create model → service ---

func (*Notifier) Wrap(m *model.SegmentMetric) *Notifier {
	return &Notifier{prefix: m.Checksum}
}

// --- Negative: context.Context — the call talks to the outside ---

func (*Notifier) Touch(ctx context.Context, m *model.SegmentMetric) error {
	return ctx.Err()
}

// --- Negative: an interface parameter — a dependency passed in ---

func (*Notifier) Check(v model.Validator, m *model.SegmentMetric) error {
	if m == nil {
		return nil
	}

	return v.Validate()
}

// --- Negative: a callback parameter ---

func (*Notifier) Each(m *model.SegmentMetric, fn func(string)) {
	fn(m.Checksum)
}

// --- Negative: the method is required by an interface of the package ---

type Differ interface {
	Differs(a, b *model.SegmentMetric) bool
}

type checksumDiffer struct{}

func (checksumDiffer) Differs(a, b *model.SegmentMetric) bool {
	return a.Checksum != b.Checksum
}

var _ Differ = checksumDiffer{}

// --- Boundary: no model parameter at all — nothing to attach the method to ---

func (*Notifier) Sum(a, b int) int {
	return a + b
}

// --- Boundary: only a slice / variadic of a model type — not a value ---

func (*Notifier) Names(ss []model.SegmentMetric) []string {
	out := make([]string, 0, len(ss))
	for i := range ss {
		out = append(out, ss[i].Checksum)
	}

	return out
}

func Join(ss ...model.SegmentMetric) string {
	return strings.Repeat("x", len(ss))
}

// --- Boundary: a model-layer interface cannot get a method ---

func Validate(a, b model.Validator) bool {
	return a == b
}

// --- Boundary: a parameter of its own package's type — not model ---

type Options struct {
	Name string
}

func OptionsName(o *Options, m *model.SegmentMetric) string {
	return o.Name + m.Checksum
}

// --- Boundary: a generic function ---

func Pick[T any](a, b T, m *model.SegmentMetric) T {
	if m.RowCount > 0 {
		return a
	}

	return b
}

// --- Boundary: a declaration without a body ---

func external(a, b *model.SegmentMetric) bool
