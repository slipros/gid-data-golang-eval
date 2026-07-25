// Boundary: a path that merely starts like the go-redis prefix is not the driver.
package dedup

import "github.com/redis/go-redis-extra-tools"

type Cache struct {
	helper *tools.Helper
}
