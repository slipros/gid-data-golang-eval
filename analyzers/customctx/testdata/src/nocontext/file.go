// Неприменимость: пакет вообще не использует context — диагностик нет.
package nocontext

type Handler struct{}

func (Handler) Serve(req string) string { return req }

// Параметр ctx, но это просто строка и контекста в пакете нет —
// для GID-188 интересны только именованные context-подобные типы;
// здесь именованного не-stdlib типа нет, диагностики нет.
func process(ctx string) string { return ctx }
