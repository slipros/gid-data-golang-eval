// Model package for the GID-278 eval.
package model

type SegmentMetric struct {
	RowCount int
	Checksum string
}

func (m *SegmentMetric) IsEmpty() bool {
	return m == nil || m.RowCount == 0
}

type Status string

const (
	StatusActive Status = "active"
	StatusDone   Status = "done"
)

type Validator interface {
	Validate() error
}
