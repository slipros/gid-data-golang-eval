// Exclusions via settings.exclude: the same violations stay silent.
package service

import "exsvc/domain/model"

type Legacy struct{}

func (*Legacy) Changed(a, b *model.SegmentMetric) bool {
	return a.RowCount != b.RowCount
}

func LegacyChanged(a, b *model.SegmentMetric) bool {
	return a.RowCount != b.RowCount
}

func (*Legacy) Reported(a, b *model.SegmentMetric) bool { // want `GID-278: method "Reported" uses neither its receiver nor the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric and drop the service that only carries it`
	return a.RowCount == b.RowCount
}
