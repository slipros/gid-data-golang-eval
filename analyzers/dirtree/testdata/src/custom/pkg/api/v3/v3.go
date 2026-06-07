// Позитив: v3 не разрешён в pkg/api кастомного дерева.
package v3 // want `GID-158: folder "v3" is not allowed in pkg/api/ \(allowed: v1, v2\); configure the tree via settings\.tree`

type API struct{}
