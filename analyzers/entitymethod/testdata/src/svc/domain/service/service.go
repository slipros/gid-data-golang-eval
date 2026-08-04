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

// --- Edge: a role suffix in the type name is not part of the entity name ---

type SessionResolver struct{}

// The entity of SessionResolver is Session: the leading part is matched.
func (r *SessionResolver) Session(ctx context.Context, id string) (Session, error) {
	return Session{}, nil
}

func (r *SessionResolver) Sessions(ctx context.Context) ([]Session, error) {
	return nil, nil
}

// A verb without the entity is still caught — the leading part is not there.
func (r *SessionResolver) Resolve(ctx context.Context) error { // want `GID-114: method name "Resolve" does not name the entity "SessionResolver"`
	return nil
}

// --- Edge: a specialised entity may be called by its more general name ---

type SessionEventOutbox struct{}

// SessionEvent is a leading part of SessionEventOutbox.
func (o *SessionEventOutbox) EnqueueSessionEvent(ctx context.Context) error {
	return nil
}

// A plural "s" is the same word: Session matches Sessions.
func (o *SessionEventOutbox) CancelPendingSessions(ctx context.Context) error {
	return nil
}

// Outbox alone is not a leading part of the entity name.
func (o *SessionEventOutbox) MarkOutboxDone(ctx context.Context) error { // want `GID-114: method name "MarkOutboxDone" does not name the entity "SessionEventOutbox"`
	return nil
}

// --- Edge: a leading part of two characters or less does not identify an entity ---

type IdCabinet struct{}

// "Id" is too short to stand for the entity, and IdCabinet is not in the name.
func (c *IdCabinet) Identify(ctx context.Context) error { // want `GID-114: method name "Identify" does not name the entity "IdCabinet"`
	return nil
}

// --- Edge: an unexported receiver is an implementation detail ---

// sessionAdapter implements a foreign interface: its method names are dictated
// by that interface, not by the type's own name — check 3 does not apply.
type sessionAdapter struct{}

func (a *sessionAdapter) EnqueueEvent(ctx context.Context) error {
	return nil
}

// The List prefix is still caught: it does not depend on the receiver at all.
func (a *sessionAdapter) ListEvents(ctx context.Context) error { // want `GID-114: drop the List prefix\. Fix: use the plural Jobs instead of ListJobs`
	return nil
}
