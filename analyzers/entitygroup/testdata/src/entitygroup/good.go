// Negative: the canonical order — type, constructor, methods;
// the entity blocks are sequential.
package entitygroup

import "context"

type Upload struct {
	id string
}

func NewUpload(id string) *Upload {
	return &Upload{id: id}
}

func (u *Upload) ID() string { return u.id }

func (u *Upload) Start(ctx context.Context) error { return nil }

type Download struct {
	id string
}

func (d *Download) ID() string { return d.id }

// Not applicable: functions without an entity do not constrain the order.
func helper(s string) string { return s }
