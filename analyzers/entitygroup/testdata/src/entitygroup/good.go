// Негатив: канонический порядок — type, конструктор, методы;
// блоки сущностей последовательны.
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

// Неприменимость: функции без сущности порядок не ограничивают.
func helper(s string) string { return s }
