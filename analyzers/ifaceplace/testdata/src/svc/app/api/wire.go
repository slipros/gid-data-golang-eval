// Eval for GID-134: control case for wire_test.go. The same
// parameter-typed-by-a-foreign-interface form, but in a non-test file — the
// exception for _test.go helpers must not leak into production code.
package api

import "svc/server/grpc"

// startServerProd has the exact shape of the startServer helper in
// wire_test.go, but this file is not a _test.go one — the violation is real
// and must still be reported.
func startServerProd(n grpc.Notifier) {} // want `GID-134: interface Notifier is declared in svc/server/grpc\. Fix: define the interface next to its consumer \(exceptions: libraries and /domain/model for service/usecase\)`
