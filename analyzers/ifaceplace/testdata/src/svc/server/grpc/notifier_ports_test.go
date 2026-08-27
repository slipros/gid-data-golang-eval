// Eval for GID-271, non-applicability class: a _test.go file is not judged
// (GID-250 keeps tests in the package of the code under test) — a port file
// in a _test.go file is silent even though its interface has exactly one
// consumer below.
package grpc

// testSink is used by the single testRecorder struct — a port file shape,
// but in a _test.go file the rule stays silent.
type testSink interface {
	Flush() error
}

// testRecorder is the only consumer of testSink.
type testRecorder struct {
	sink testSink
}
