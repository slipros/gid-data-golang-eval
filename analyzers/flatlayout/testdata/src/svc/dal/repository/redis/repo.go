// Positive: a grouping subpackage by storage type is forbidden.
package redis // want `GID-138: package "svc/dal/repository/redis"\. Fix: grouping subpackages in /dal/repository are forbidden, keep layer entities at its root`

type Cache struct{}
