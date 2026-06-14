// Stub of a pb package for the settings.exclude eval.
package orderpb

import "google.golang.org/grpc"

type OrderClient struct{ cc *grpc.ClientConn }

func (c *OrderClient) Get(id string) string { return id }
