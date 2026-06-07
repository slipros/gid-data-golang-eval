// Позитив: чужая папка глубже internal — generic-сообщение без подсказки.
package redis // want `GID-158: папка "redis" не разрешена в internal/dal/ \(разрешены: entity, repository\); дерево настраивается через settings\.tree`

type Conn struct{}
