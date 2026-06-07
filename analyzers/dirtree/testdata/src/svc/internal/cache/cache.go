// Позитив: чужая папка в internal/ — подсказка про service/usecase.
package cache // want `GID-158: folder "cache" is not allowed in internal/ \(allowed: app, client, dal, domain, event, metric, server\); perhaps it should be a service or usecase; configure the tree via settings\.tree`

type Cache struct{}
