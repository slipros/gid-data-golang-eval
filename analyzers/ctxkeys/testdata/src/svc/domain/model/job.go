// Негатив GID-166: helper'ы сущности Job в файле её объявления,
// ключ — из context.go.
package model

import "context"

type Job struct {
	ID string
}

func ContextWithJob(ctx context.Context, j Job) context.Context {
	return context.WithValue(ctx, JobKey, j)
}

func JobFromContext(ctx context.Context) (Job, bool) {
	j, ok := ctx.Value(JobKey).(Job)
	return j, ok
}

// Позитив GID-166: helper сущности Token, объявленной в session.go.
func TokenFromContext(ctx context.Context) (Token, bool) { // want `GID-166: helper "TokenFromContext" must live in the same file as the "Token" entity it stores into / reads from ctx`
	t, ok := ctx.Value(TraceIDKey).(Token)
	return t, ok
}
