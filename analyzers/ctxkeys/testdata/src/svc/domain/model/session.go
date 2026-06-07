// Негатив GID-166: сущность и её ctx-helper'ы в одном файле; ключи —
// в файле типа ContextKey (context.go).
package model

import "context"

type Session struct {
	ID string
}

type Token struct {
	Value string
}

func ContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, SessionKey, s)
}

func SessionFromContext(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(SessionKey).(Session)
	return s, ok
}

// Позитив GID-167: значение ContextKey вне файла объявления типа.
const LegacySessionKey ContextKey = "legacy_session" // want `GID-167: значения ContextKey находятся рядом с объявлением типа ContextKey \(в одном файле\)`

// Позитив GID-166: имя без паттерна <Name>FromContext.
func JobFromContextLegacy(ctx context.Context) (Job, bool) { // want `GID-166: функция "JobFromContextLegacy" достаёт данные из ctx — она публична и именуется <Name>FromContext`
	j, ok := ctx.Value(JobKey).(Job)
	return j, ok
}