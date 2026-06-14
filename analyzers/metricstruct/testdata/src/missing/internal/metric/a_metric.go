// The file with the smallest name — the report that the metric package does
// not declare a Prometheus aggregator is deterministically placed here.
package metric // want `GID-174: the metric package must declare a metrics aggregator: struct Prometheus with a Register method\. Fix: add it`

// HTTPRequests — some metric, but without a Prometheus aggregator.
var HTTPRequests int
