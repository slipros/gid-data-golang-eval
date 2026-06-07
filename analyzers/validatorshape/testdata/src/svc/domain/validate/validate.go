// Eval GID-213: форма валидатора в слое validate.
package validate

import "context"

// jobReq — тип запроса для валидаторов.
type jobReq struct{ ID string }

// jobOpt — дополнительный параметр для граничного кейса.
type jobOpt struct{}

// --- Позитивный класс: нарушения ---

// Нет метода Validate вовсе.
type CreateJob struct{} // want `GID-213: валидатор "CreateJob" обязан иметь метод Validate\(ctx context.Context, req \*T\) error`

// Validate без ctx первым параметром.
type UpdateJob struct{} // want `GID-213: валидатор "UpdateJob" обязан иметь метод Validate\(ctx context.Context, req \*T\) error`

func (v *UpdateJob) Validate(req *jobReq) error { return nil }

// Validate возвращает (bool, error), а не один error.
type DeleteJob struct{} // want `GID-213: валидатор "DeleteJob" обязан иметь метод Validate\(ctx context.Context, req \*T\) error`

func (v *DeleteJob) Validate(ctx context.Context, req *jobReq) (bool, error) { return false, nil }

// --- Негативный класс: корректный код ---

// Корректный валидатор (pointer-receiver) — ок.
type ListJobs struct{}

func (v *ListJobs) Validate(ctx context.Context, req *jobReq) error { return nil }

// Тип-настройки (*Options) — не валидатор, не флагается.
type ListJobsOptions struct{ Limit int }

// Неэкспортируемый struct — не флагается.
type helper struct{} //nolint:unused

// --- Граничный класс ---

// value-receiver — ок.
type GetJob struct{}

func (v GetJob) Validate(ctx context.Context, req *jobReq) error { return nil }

// Два параметра после ctx — ctx + единственный error достаточны, ок.
type PatchJob struct{}

func (v PatchJob) Validate(ctx context.Context, req *jobReq, opt jobOpt) error { return nil }
