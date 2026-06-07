// Файл с наименьшим именем — сюда детерминированно ставится репорт о
// том, что пакет metric не объявляет агрегатор Prometheus.
package metric // want `GID-174: the metric package must declare a metrics aggregator: struct Prometheus with a Register method\. Fix: add it`

// HTTPRequests — какая-то метрика, но без агрегатора Prometheus.
var HTTPRequests int
