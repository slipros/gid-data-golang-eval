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
const LegacySessionKey ContextKey = "legacy_session" // want `GID-167: ContextKey values must be declared next to the ContextKey type declaration \(same file\)`

// Позитив GID-166: имя без паттерна <Name>FromContext.
func JobFromContextLegacy(ctx context.Context) (Job, bool) { // want `GID-166: function "JobFromContextLegacy" reads data from ctx\. Fix: make it public and name it <Name>FromContext`
	j, ok := ctx.Value(JobKey).(Job)
	return j, ok
}