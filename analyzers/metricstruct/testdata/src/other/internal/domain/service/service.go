// A package outside the metric path: the rule does not apply even when a
// Prometheus without Register is present (not applicable).
package service

// Prometheus without Register — but the path is not metric, so no diagnostic.
type Prometheus struct {
	HTTP int
}
