// The path ends with metrics — package-name violation (report on the package clause).
package metric // want `GID-174: the metrics package must be named metric, not metrics\. Fix: rename it to metric`

// Prometheus is present and correct here, but the metrics path is still a violation.
type Prometheus struct {
	HTTP int
}

// Register is correct.
func (p Prometheus) Register() error { return nil }
