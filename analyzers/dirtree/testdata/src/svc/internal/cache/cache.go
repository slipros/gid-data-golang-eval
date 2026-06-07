// Позитив: чужая папка в internal/ — подсказка про service/usecase.
package cache // want `GID-158: папка "cache" не разрешена в internal/ \(разрешены: app, client, dal, domain, event, metric, server\) — возможно, это должен быть service или usecase; дерево настраивается через settings\.tree`

type Cache struct{}
