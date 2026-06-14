// Positive (boundary): build/ is allowed only for a repository, not for a service.
package build // want `GID-138: package "svc/domain/service/build"\. Fix: grouping subpackages in /domain/service are forbidden, keep layer entities at its root`

func Helper() {}
