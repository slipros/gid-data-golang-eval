// Eval for GID-114 (service): the root package /domain/service is in scope.
package service

import "context"

type Session struct{ ID string }

// S — a single-letter "entity": check 3 does not apply (a utility name).
type S struct{}

// --- Positive ---

func (s *Session) ListSessions(ctx context.Context) ([]Session, error) { // want `GID-114: drop the List prefix\. Fix: use the plural Jobs instead of ListJobs`
	return nil, nil
}

func (s *Session) SessionByID(ctx context.Context, id string) (Session, error) { // want `GID-114: drop the ByID suffix\. Fix: use Job\(ctx, id\) instead of JobByID`
	return Session{}, nil
}

// --- Negative ---

func (s *Session) Session(ctx context.Context, id string) (Session, error) {
	return Session{}, nil
}

func (s *Session) Sessions(ctx context.Context) ([]Session, error) {
	return nil, nil
}

// --- Edge: the single-letter receiver S — the entity name is not checked ---

// The method name lacks "S", but the entity is a utility one (len <= 2) — no diagnostic.
func (x *S) Touch(ctx context.Context) error {
	return nil
}

// The List prefix is still caught — it does not depend on the entity name length.
func (x *S) ListAll(ctx context.Context) error { // want `GID-114: drop the List prefix\. Fix: use the plural Jobs instead of ListJobs`
	return nil
}
