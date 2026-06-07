// Eval GID-166/167: форма ctx-helper'ов и ключей в model.
// Канон: публичный type ContextKey string, все значения — в этом файле,
// string в snake_case.
package model

import "context"

type ContextKey string

const (
	// Негатив: канонические значения.
	UserIDKey  ContextKey = "user_id"
	TraceIDKey ContextKey = "trace_id"
	SessionKey ContextKey = "session"
	JobKey     ContextKey = "job"

	// Позитив: не snake_case.
	BadCamelKey ContextKey = "UserID" // want `GID-167: ContextKey value must be a snake_case string, got "UserID"`
	BadDashKey  ContextKey = "user-id" // want `GID-167: ContextKey value must be a snake_case string, got "user-id"`
)

type secretKey string

// --- GID-166: кладёт в ctx, но имя не ContextWith<Name> ---

func WithUserID(ctx context.Context, id string) context.Context { // want `GID-166: function "WithUserID" stores data in ctx\. Fix: make it public and name it ContextWith<Name>`
	return context.WithValue(ctx, UserIDKey, id)
}

// Граничный кейс: приватный helper — ContextWith требует публичности.
func contextWithTrace(ctx context.Context, id string) context.Context { // want `GID-166: function "contextWithTrace" stores data in ctx\. Fix: make it public and name it ContextWith<Name>`
	return context.WithValue(ctx, TraceIDKey, id)
}

// --- GID-166: достаёт из ctx, но имя не <Name>FromContext ---

func GetUserID(ctx context.Context) (string, bool) { // want `GID-166: function "GetUserID" reads data from ctx\. Fix: make it public and name it <Name>FromContext`
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

// --- GID-167: ключ не типа ContextKey ---

func ContextWithSecret(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, secretKey("secret"), s) // want `GID-167: context key must be the public type ContextKey \(type ContextKey string\), not "secretKey"`
}

// Граничный кейс: сырой string-ключ.
func ContextWithRaw(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, "raw", s) // want `GID-167: context key must be the public type ContextKey \(type ContextKey string\), not a raw value`
}

// --- Негатив: канонические helper'ы ---

func ContextWithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

// --- Неприменимость: функции без работы с ctx-значениями ---

func Normalize(id string) string { return id }
