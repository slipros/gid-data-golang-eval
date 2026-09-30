// Non-applicability: a subpackage of service (convert) is not a layer root.
package convert

import "svc/domain/model"

func Changed(a, b *model.SegmentMetric) bool {
	return a.Checksum != b.Checksum
}
