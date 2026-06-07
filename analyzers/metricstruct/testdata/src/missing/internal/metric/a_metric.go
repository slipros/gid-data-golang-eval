// Файл с наименьшим именем — сюда детерминированно ставится репорт о
// том, что пакет metric не объявляет агрегатор Prometheus.
package metric // want `GID-174: пакет metric объявляет агрегатор метрик — struct Prometheus с методом Register`

// HTTPRequests — какая-то метрика, но без агрегатора Prometheus.
var HTTPRequests int
