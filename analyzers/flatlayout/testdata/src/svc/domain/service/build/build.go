// Позитив (граница): build/ разрешён только у репозитория, у сервиса — нет.
package build // want `GID-138: package "svc/domain/service/build"\. Fix: grouping subpackages in /domain/service are forbidden, keep layer entities at its root`

func Helper() {}
