// Позитив (граница): build/ разрешён только у репозитория, у сервиса — нет.
package build // want `GID-138: пакет "svc/domain/service/build" — группирующие подпакеты в /domain/service запрещены, сущности слоя живут в его корне`

func Helper() {}
