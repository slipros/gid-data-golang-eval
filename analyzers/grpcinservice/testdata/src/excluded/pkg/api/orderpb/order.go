// Stub pb-пакета для eval settings.exclude.
package orderpb

import "google.golang.org/grpc"

type OrderClient struct{ cc *grpc.ClientConn }

func (c *OrderClient) Get(id string) string { return id }
