// Positive: a foreign folder in internal/ — the hint about service/usecase.
package cache // want `GID-158: folder "cache" is not allowed in internal/ \(allowed: app, client, dal, domain, event, job, metric, schedule, server\); perhaps it should be a service or usecase; configure the tree via settings\.tree`

type Cache struct{}
