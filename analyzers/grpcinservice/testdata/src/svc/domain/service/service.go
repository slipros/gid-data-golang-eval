// Eval for GID-160: a service calls gRPC through a repository.
package service

import (
	"google.golang.org/grpc" // want `GID-160: direct import of google\.golang\.org/grpc in the domain layer is forbidden\. Fix: call gRPC through a repository \(exceptions: nolint or settings\.exclude\)`

	"svc/pkg/api/orderpb" // want `GID-160: importing the gRPC package "svc/pkg/api/orderpb" in the domain layer is forbidden\. Fix: call gRPC through a repository`
)

type Order struct {
	conn   *grpc.ClientConn
	client *orderpb.OrderClient
}
