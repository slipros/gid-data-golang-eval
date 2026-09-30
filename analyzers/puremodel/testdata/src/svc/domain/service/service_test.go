package service

import "svc/domain/model"

// --- Non-applicability: a _test.go file — a double mirrors the interface it fakes ---

type fakeDiffer struct{}

func (fakeDiffer) Compare(a, b *model.SegmentMetric) bool {
	return a.Checksum == b.Checksum
}
