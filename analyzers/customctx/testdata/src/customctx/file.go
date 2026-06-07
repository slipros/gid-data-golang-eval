// Eval для GID-188 (запрет кастомных context-типов).
package customctx

import (
	"context"
	"time"
)

// --- Класс 1: позитивные (нарушения) ---

// Кейс 2: interface, встраивающий context.Context.
type MyContext interface { // want `GID-188: кастомный context-тип MyContext запрещён — передавайте context\.Context и кладите данные через context\.WithValue \(хелперы в /domain/model — GID-165/166\)`
	context.Context
	Extra() string
}

// Кейс 1: struct с полным набором методов context.Context.
type CtxStruct struct{} // want `GID-188: кастомный context-тип CtxStruct запрещён`

func (CtxStruct) Deadline() (time.Time, bool) { return time.Time{}, false }
func (CtxStruct) Done() <-chan struct{}       { return nil }
func (CtxStruct) Err() error                  { return nil }
func (CtxStruct) Value(key any) any           { return nil }

// Кейс 3: параметр ctx — кастомный context-тип.
func useCustom(ctx MyContext) {} // want `GID-188: параметр ctx имеет тип .*MyContext — используйте context\.Context`

// --- Класс 2: негативные (чистый код) ---

// Параметр ctx правильного типа.
func good(ctx context.Context) { _ = ctx }

// struct с одним методом Done — method set НЕ покрывает context.Context.
type PartialCtx struct{}

func (PartialCtx) Done() <-chan struct{} { return nil }

// --- Класс 3: граничные ---

// interface { context.Context } — embedding матчится один раз (одна диагностика).
type OnlyEmbed interface { // want `GID-188: кастомный context-тип OnlyEmbed запрещён`
	context.Context
}

// Тип с методами Deadline/Done/Err/Value, но другими сигнатурами — не context.Context.
type FakeCtx struct{}

func (FakeCtx) Deadline() string  { return "" }
func (FakeCtx) Done() bool        { return false }
func (FakeCtx) Err() string       { return "" }
func (FakeCtx) Value(key int) int { return 0 }

// Параметр с именем ctx, но stdlib-типом — не нарушение.
func boundaryGood(ctx context.Context, other FakeCtx) { _ = ctx; _ = other }
