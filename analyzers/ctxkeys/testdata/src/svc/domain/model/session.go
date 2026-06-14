// Negative GID-166: the entity and its ctx helpers in one file; the keys are
// in the file of the ContextKey type (context.go).
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

// Positive GID-167: a ContextKey value outside the type declaration file.
const LegacySessionKey ContextKey = "legacy_session" // want `GID-167: ContextKey values must be declared next to the ContextKey type declaration \(same file\)`

// Positive GID-166: a name not following the <Name>FromContext pattern.
func JobFromContextLegacy(ctx context.Context) (Job, bool) { // want `GID-166: function "JobFromContextLegacy" reads data from ctx\. Fix: make it public and name it <Name>FromContext`
	j, ok := ctx.Value(JobKey).(Job)
	return j, ok
}