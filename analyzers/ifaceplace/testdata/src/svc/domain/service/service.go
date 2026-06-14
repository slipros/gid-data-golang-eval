// Eval for GID-134 (interface-near-consumer). The consumer is the
// /domain/service layer.
package service

import (
	"io"

	"example.com/extlib"
	"svc/domain/model"
	"svc/server/grpc"
)

// LocalRepository — an interface declared in this same package next to the
// consumer. Its use is the norm.
type LocalRepository interface {
	Job(id string) (model.Job, error)
}

// --- Positive class: an interface from another "own" service package ---

// A struct field: an interface from a foreign server package.
type Service struct {
	notifier grpc.Notifier // want `GID-134: interface Notifier is declared in svc/server/grpc\. Fix: define the interface next to its consumer \(exceptions: libraries and /domain/model for service/usecase\)`
	local    LocalRepository
}

// A function parameter: an interface from a foreign server package.
func (s *Service) Register(n grpc.Notifier) {} // want `GID-134: interface Notifier is declared in svc/server/grpc`

// A function result: an interface from a foreign server package.
func (s *Service) Notifier() grpc.Notifier { return nil } // want `GID-134: interface Notifier is declared in svc/server/grpc`

// --- Negative class: clean code ---

// An interface from the model layer at a service consumer — OK.
func (s *Service) WithRepo(r model.JobRepository) {}

// An interface from the same package — OK.
func (s *Service) WithLocal(l LocalRepository) {}

// A stdlib library interface (io.Reader) — OK.
func (s *Service) Read(r io.Reader) {}

// An external library interface — OK.
func (s *Service) Encode(e extlib.Encoder) {}

// --- Inapplicability class ---

// error — untouched (no declaring package).
func (s *Service) Do() error { return nil }

// An anonymous interface — not named, untouched.
func (s *Service) Anon(x interface{ Foo() }) {}

// any / interface{} — untouched.
func (s *Service) Any(v any) {}

// Non-interface types (struct, string) — untouched.
func (s *Service) Plain(j model.Job, name string) {}
