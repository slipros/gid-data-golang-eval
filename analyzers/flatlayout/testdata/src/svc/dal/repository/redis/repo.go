// Позитив: группирующий подпакет по типу хранилища запрещён.
package redis // want `GID-138: пакет "svc/dal/repository/redis" — группирующие подпакеты в /dal/repository запрещены, сущности слоя живут в его корне`

type Cache struct{}
