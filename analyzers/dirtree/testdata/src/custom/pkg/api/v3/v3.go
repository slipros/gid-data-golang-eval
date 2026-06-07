// Позитив: v3 не разрешён в pkg/api кастомного дерева.
package v3 // want `GID-158: папка "v3" не разрешена в pkg/api/ \(разрешены: v1, v2\); дерево настраивается через settings\.tree`

type API struct{}
