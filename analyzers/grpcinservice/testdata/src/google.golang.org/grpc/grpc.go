// Stub of google.golang.org/grpc for eval.
package grpc

type ClientConn struct{}

func Dial(target string) (*ClientConn, error) { return &ClientConn{}, nil }
