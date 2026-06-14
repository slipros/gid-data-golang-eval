// Eval GID-166/167: the shape of ctx helpers and keys in model.
// Canon: a public type ContextKey string, all values in this file,
// strings in snake_case.
package model

import "context"

type ContextKey string

const (
	// Negative: canonical values.
	UserIDKey  ContextKey = "user_id"
	TraceIDKey ContextKey = "trace_id"
	SessionKey ContextKey = "session"
	JobKey     ContextKey = "job"

	// Positive: not snake_case.
	BadCamelKey ContextKey = "UserID" // want `GID-167: ContextKey value must be a snake_case string, got "UserID"`
	BadDashKey  ContextKey = "user-id" // want `GID-167: ContextKey value must be a snake_case string, got "user-id"`
)

type secretKey string

// --- GID-166: stores into ctx, but the name is not ContextWith<Name> ---

func WithUserID(ctx context.Context, id string) context.Context { // want `GID-166: function "WithUserID" stores data in ctx\. Fix: make it public and name it ContextWith<Name>`
	return context.WithValue(ctx, UserIDKey, id)
}

// Edge case: a private helper — ContextWith requires being public.
func contextWithTrace(ctx context.Context, id string) context.Context { // want `GID-166: function "contextWithTrace" stores data in ctx\. Fix: make it public and name it ContextWith<Name>`
	return context.WithValue(ctx, TraceIDKey, id)
}

// --- GID-166: reads from ctx, but the name is not <Name>FromContext ---

func GetUserID(ctx context.Context) (string, bool) { // want `GID-166: function "GetUserID" reads data from ctx\. Fix: make it public and name it <Name>FromContext`
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

// --- GID-167: a key not of the ContextKey type ---

func ContextWithSecret(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, secretKey("secret"), s) // want `GID-167: context key must be the public type ContextKey \(type ContextKey string\), not "secretKey"`
}

// Edge case: a raw string key.
func ContextWithRaw(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, "raw", s) // want `GID-167: context key must be the public type ContextKey \(type ContextKey string\), not a raw value`
}

// --- Negative: canonical helpers ---

func ContextWithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

// --- Not applicable: functions that do not work with ctx values ---

func Normalize(id string) string { return id }
