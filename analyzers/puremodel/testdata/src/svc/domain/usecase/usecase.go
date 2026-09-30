// Eval for GID-278 (puremodel): the usecase root is in scope as well.
package usecase

import "svc/domain/model"

type Refresh struct{}

func (Refresh) Stale(a, b *model.SegmentMetric) bool { // want `GID-278: method "Stale" uses neither its receiver nor the package, it is pure logic over model.SegmentMetric\. Fix: make it a public method of model.SegmentMetric and drop the service that only carries it`
	return a.RowCount != b.RowCount
}
