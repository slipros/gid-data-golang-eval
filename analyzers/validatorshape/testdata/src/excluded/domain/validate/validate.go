// Eval GID-213: тип из settings.exclude не считается валидатором.
package validate

import "context"

type jobReq struct{ ID string }

// Числится в settings.exclude как "HealthCheck" — без Validate не флагается.
type HealthCheck struct{}

// Корректный валидатор — не флагается в любом случае.
type Ping struct{}

func (v *Ping) Validate(ctx context.Context, req *jobReq) error { return nil }

// Не в exclude и без корректного Validate — флагается, проверяет, что
// исключение не глушит весь пакет.
type Echo struct{} // want `GID-213: validator "Echo" must have a Validate\(ctx context.Context, req \*T\) error method\. Fix: add it`
