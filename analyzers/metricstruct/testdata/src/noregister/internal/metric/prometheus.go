// Prometheus exists but has no Register method — violation (check 3).
package metric

// Prometheus — a struct without a Register method.
type Prometheus struct { // want `GID-174: struct Prometheus must have a Register method\. Fix: add it`
	HTTP int
}

// Collect — an unrelated method, not Register.
func (p Prometheus) Collect() {}
