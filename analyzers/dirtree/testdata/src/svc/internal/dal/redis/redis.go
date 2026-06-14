// Positive: a foreign folder deeper than internal — a generic message without the hint.
package redis // want `GID-158: folder "redis" is not allowed in internal/dal/ \(allowed: entity, repository\); configure the tree via settings\.tree`

type Conn struct{}
