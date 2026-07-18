// Eval for GID-216 (boundary): a package with segments event and consumer
// nested under a DIFFERENT layer (server) — this is a server-layer package
// that happens to use the words event/consumer, not the module's event
// layer. Before the pathseg.HasLayer fix, pathseg.Contains(pkgPath, "event")
// matched "event" anywhere in the path and falsely put this package in
// event-consumer scope. With HasLayer (anchored to the module root, which
// here is "server", not "event") the rule does not apply — no diagnostic.
package consumer

// OrderConsumer — a constructor without a logger must NOT be flagged here:
// the package is not in the event layer (it is nested under server/grpc).
type OrderConsumer struct{}

func NewOrderConsumer() *OrderConsumer {
	return &OrderConsumer{}
}
