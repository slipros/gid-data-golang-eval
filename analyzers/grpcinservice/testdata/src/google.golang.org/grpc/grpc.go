// Stub google.golang.org/grpc для eval.
package grpc

type ClientConn struct{}

func Dial(target string) (*ClientConn, error) { return &ClientConn{}, nil }
