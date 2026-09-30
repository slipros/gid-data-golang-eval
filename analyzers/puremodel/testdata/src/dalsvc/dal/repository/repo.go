// Non-applicability: /dal/repository is outside the rule's layers.
package repository

import "dalsvc/domain/model"

type Repo struct{}

func (*Repo) Changed(a, b *model.SegmentMetric) bool {
	return a.RowCount != b.RowCount
}
