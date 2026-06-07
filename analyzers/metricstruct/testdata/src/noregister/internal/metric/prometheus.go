// Prometheus есть, но без метода Register — нарушение (проверка 3).
package metric

// Prometheus — struct без метода Register.
type Prometheus struct { // want `GID-174: struct Prometheus must have a Register method\. Fix: add it`
	HTTP int
}

// Collect — посторонний метод, не Register.
func (p Prometheus) Collect() {}
