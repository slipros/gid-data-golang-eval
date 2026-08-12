// Stub of google.golang.org/grpc for eval: the call-option type is the marker
// GID-266 recognises a gRPC client call by.
package grpc

type CallOption interface {
	before() error
}

type EmptyCallOption struct{}

func (EmptyCallOption) before() error { return nil }

type waitForReadyOption struct {
	EmptyCallOption

	ready bool
}

func WaitForReady(ready bool) CallOption { return waitForReadyOption{ready: ready} }
