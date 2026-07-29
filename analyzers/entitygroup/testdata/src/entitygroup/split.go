// Eval for GID-157 — free declarations tearing the entity block apart.
package entitygroup

type Session struct {
	id string
}

func NewSession(id string) *Session {
	return &Session{id: id}
}

func (s *Session) ID() string { return s.id }

// Positive: a free helper between the entity's methods.
func sessionKey(id string) string { // want `GID-157: "sessionKey" splits the "Session" entity block\. Fix: move it above the first type or below the entity's last method`
	return "session:" + id
}

// Positive: a non-struct type inside the block.
type sessionState string // want `GID-157: "sessionState" splits the "Session" entity block\. Fix: move it above the first type or below the entity's last method`

func (s *Session) Key() string { return sessionKey(s.id) }

// Positive: a second constructor below the methods — every constructor of an
// entity belongs under its type declaration.
func NewSessionFromID(id string) (*Session, error) { // want `GID-157: constructor "NewSessionFromID" sits below the methods of "Session"\. Fix: keep every constructor of an entity together under its type declaration, above the methods`
	return &Session{id: id}, nil
}
