// Eval for GID-134: a _test.go file is not exempted wholesale. A wiring test
// (mirrors resource-registry yandex_audience_wire_test.go, incident
// 2026-08-06) has no choice about a helper's parameter/result types — they
// are dictated by the production constructor it wires up — but a struct type
// the test declares on its own is a free choice and stays in scope.
package api

import "svc/server/grpc"

// --- Non-applicability class: usage forced by a production constructor ---

// startServer mirrors a wire-test helper that boots a real handler: its
// parameter is typed with the consumer-side interface (grpc.Notifier) only
// because that is what the production constructor it calls demands. Not
// flagged: a helper's parameters in a _test.go file are a use, not a
// declaration.
func startServer(n grpc.Notifier) {}

// notifierFixture mirrors a wire-test helper handing back a value typed by
// the production constructor's own result. Not flagged for the same reason.
func notifierFixture() grpc.Notifier { return nil }

// --- Boundary class: a declaration in the same _test.go file stays in scope ---

// stubServer is a struct the test declared on its own — nothing forced this
// field's type, so GID-134 keeps catching it even inside a _test.go file.
type stubServer struct {
	notifier grpc.Notifier // want `GID-134: interface Notifier is declared in svc/server/grpc\. Fix: define the interface next to its consumer \(exceptions: libraries and /domain/model for service/usecase\)`
}
