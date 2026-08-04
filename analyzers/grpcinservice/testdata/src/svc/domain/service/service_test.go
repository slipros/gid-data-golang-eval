// Non-applicability: a _test.go file is not judged. A test lives in the same
// package (GID-250), and a double of a consumer-side client interface has to
// repeat its signatures — including the types of the generated client it
// stands in for.
package service

import (
	"testing"

	"google.golang.org/grpc"
)

// fakeConn doubles the gRPC connection the service under test is wired to.
type fakeConn struct{ conn *grpc.ClientConn }

func TestOrder(t *testing.T) {
	if (&fakeConn{}).conn != nil {
		t.Fatal("unexpected conn")
	}
}
